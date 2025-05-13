package main

import (
	"context"
	"flag"

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/config"
	"github.com/alnah/fla/logger"
)

func main() {
	/********* Setup logger *********/
	log := logger.New()

	/********* Setup config *********/
	path, err := config.Path()
	if err != nil {
		log.Error("retrieving config path", "error", err.Error())
		return
	}
	cfg, err := config.Load(log, path)
	if err != nil {
		log.Error("loading config", "error", err.Error())
		return
	}
	cfg.BindFlags()
	flag.Parse()
	if err := cfg.Validate(); err != nil {
		log.Error("validating config", "error", err.Error())
		return
	}
	if err := cfg.EnsureIODirs(); err != nil {
		log.Error("ensuring application directories", "error", err.Error())
		return
	}

	/********* Log config *********/
	cfg.Log.Info("config",
		"language", cfg.Language,
		"input", cfg.Input,
		"output", cfg.Output,
		"ai_completion_timeout", cfg.AI.Timeout.Completion,
		"ai_audio_timeout", cfg.AI.Timeout.Audio,
		"ai_transcription_timeout", cfg.AI.Timeout.Transcription,
	)

	/******** Demo *********/
	ctx, cancel := context.WithTimeout(context.Background(), cfg.AI.Timeout.Completion)
	defer cancel()

	chat, err := ai.NewChatClient(
		ai.WithProvider(ai.ProviderAnthropic),
		ai.WithURL(ai.URLChatCompletionAnthropic),
		ai.WithAPIKey(ai.EnvAnthropicAPIKey),
		ai.WithModel(ai.AIModelCostOptimizedAnthropic),
		ai.WithContext(ctx),
		ai.WithLogger(log),
		ai.WithMessages(ai.Messages{
			ai.Message{
				Role:    ai.RoleUser,
				Content: "Comment vas-tu ?",
			},
		}),
	)
	if err != nil {
		log.Error("new chat client failed", "error", err.Error())
		return
	}
	completion, err := chat.Do()
	if err != nil {
		log.Error("completion failed", "error", err.Error())
		return
	}
	log.Info("completion succeed", "content", completion.String())
}
