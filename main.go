package main

import (
	"flag"
	"os"

	"github.com/alnah/fla/ai"
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
		log.Error("validating config", "error", err.Error())
		return
	}
	if err := cfg.EnsureDirs(); err != nil {
		log.Error("ensuring application directories", "error", err.Error())
	}

	/********* Log config *********/
	log.Info("config",
		"language", cfg.Language,
		"input", cfg.Input,
		"output", cfg.Output,
		"ai_completion_timeout", cfg.AI.Timeout.Completion,
		"ai_audio_timeout", cfg.AI.Timeout.Audio,
		"ai_transcription_timeout", cfg.AI.Timeout.Transcription,
	)
}
