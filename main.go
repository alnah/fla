package main

import (
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

	transcript, err := ai.NewSTTClient(
		ai.WithLogger[*ai.STTClient](log),
		ai.WithProvider[*ai.STTClient](ai.ProviderElevenLabs),
		ai.WithURL[*ai.STTClient](ai.URLTranscriptElevenLabs),
		ai.WithAPIKey[*ai.STTClient](ai.EnvElevenLabsAPIKey),
		ai.WithModel[*ai.STTClient](ai.AIModelTranscriptElevenLabs),
		ai.WithFilePath("tts.mp3"),
		ai.WithLanguage("fr"),
	)
	if err != nil {
		log.Error("transcript client", "error", err.Error())
		return
	}
	result, err := transcript.Transcript()
	if err != nil {
		log.Error("transcript doer", "error", err.Error())
		return
	}
	log.Info("transcript", "result", result.Content())
}
