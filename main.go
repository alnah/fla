package main

import (
	"context"
	"log/slog"
	"os"

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/config"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/storage/cache"
)

func main() {
	/********* Setup logger *********/
	log := logger.NewSlogger(os.Stdout, true, slog.LevelDebug)
	cfg, err := config.New(config.WithLogger(log)).Load()
	if err != nil {
		log.Error("loading configuration", "error", err.Error())
		return
	}
	if err := cfg.Save(); err != nil {
		log.Error("updating configuratio", "error", err.Error())
	}

	ctx, cancel := cfg.ChatContext(context.Background())
	defer cancel()
	/********* Setup cache store *********/
	rc, err := cache.NewRedisCache().
		WithAddress(cfg.Cache.Address).
		WithPassword(cfg.Cache.Password).
		WithLogger(log).
		Build()

	if err != nil {
		log.Error("redis cache", "error", err.Error())
		return
	}
	err = rc.Set(context.Background(), "hello", "world", 0)
	if err != nil {
		log.Error("redis cache", "error", err.Error())
		return
	}
	val, err := rc.Get(context.Background(), "hello")
	if err != nil {
		log.Error("redis cache", "error", err.Error())
		return
	}
	log.Info("redis", "result", val.String())

	/********* Setup chat client *********/
	chat, err := ai.NewChatClient(
		ai.WithLogger[*ai.ChatClient](log),
		ai.WithContext[*ai.ChatClient](ctx),
		ai.WithProvider[*ai.ChatClient](ai.ProviderOpenAI),
		ai.WithURL[*ai.ChatClient](ai.URLChatOpenAI),
		ai.WithAPIKey[*ai.ChatClient](cfg.AI.APIKey.OpenAI),
		ai.WithModel[*ai.ChatClient](ai.ModelCheapOpenAI),
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
