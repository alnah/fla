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
	cfg, err := config.NewLoader(config.WithLogger(log)).Load()
	if err != nil {
		log.Error("configuration loading", "error", err.Error())
		return
	}

	ctx, cancel := cfg.ChatContext(context.Background())
	defer cancel()

	/********* Setup chat client *********/
	chat, err := ai.NewChatClient(
		ai.WithLogger[*ai.ChatClient](log),
		ai.WithContext[*ai.ChatClient](ctx),
		ai.WithProvider[*ai.ChatClient](ai.ProviderAnthropic),
		ai.WithURL[*ai.ChatClient](ai.URLChatAnthropic),
		ai.WithAPIKey[*ai.ChatClient](cfg.APIKey.Anthropic),
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
		log.Error("chat client setup", "error", err.Error())
		return
	}

	/********* Call completion *********/
	completion, err := chat.Completion()
	if err != nil {
		log.Error("chat completion", "error", err.Error())
		return
	}
	log.Info("chat completion", "content", completion.Content())
}
