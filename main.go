package main

import (
	"context"
	"flag"
	"os"

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/aiclient/openai"
	"github.com/alnah/fla/config"
	"github.com/alnah/fla/logger"
)

func main() {
	/********* Setup logger *********/
	log := logger.New()

	/********* Validate env *********/
	required := []string{
		ai.EnvAPIKeyOpenAI,
		ai.EnvAPIKeyAnthropic,
		ai.EnvAPIKeyElevenLabs,
	}
	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Warn("missing environment variable", "key", key)
		}
	}

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
		cfg.Log.Error("validating config", "error", err.Error())
		return
	}
	if err := cfg.EnsureDirs(); err != nil {
		cfg.Log.Error("ensuring application directories", "error", err.Error())
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

	completion, err := openai.NewChat(openai.WithContext[*openai.Chat](ctx), openai.WithLogger[*openai.Chat](cfg.Log)).
		SetSystem("Tu es un professeur de français.").
		AddMessage(ai.RoleUser, "Tu dois écrire un texte de niveau B1 FLE. Retourne uniquement le texte.").
		Completion()
	if err != nil {
		cfg.Log.Error("completion failed", "error", err.Error())
		return
	}
	cfg.Log.Info("completion succeed", "content", completion.Content())
}
