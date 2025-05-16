package main

import (
	"context"

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
	cfg, err := config.Load(log, path.DirPath, path.FileName)
	if err != nil {
		log.Error("loading config", "error", err.Error())
		return
	}
	if err := cfg.EnsureIODirs(); err != nil {
		log.Error("ensuring application directories", "error", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.AI.Timeout.Chat)
	defer cancel()

	chat, err := ai.NewChatClient(
		ai.WithContext[*ai.ChatClient](ctx),
		ai.WithLogger[*ai.ChatClient](log),
		ai.WithProvider[*ai.ChatClient](ai.ProviderAnthropic),
		ai.WithURL[*ai.ChatClient](ai.URLChatAnthropic),
		ai.WithAPIKey[*ai.ChatClient](ai.APIKeyEnvAnthropic),
		ai.WithModel[*ai.ChatClient](ai.ModelCheapAnthropic),
		ai.WithMaxTokens(100),
		ai.WithTemperature(0.2),
		ai.WithMessages(ai.Messages{
			ai.Message{
				Role:    ai.RoleUser,
				Content: "say hi",
			},
		}),
	)
	if err != nil {
		log.Error("chat", "error", err.Error())
		return
	}
	completion, err := chat.Completion()
	if err != nil {
		log.Error("completion", "error", err.Error())
		return
	}
	log.Info("completion", "content", completion.Content())
}
