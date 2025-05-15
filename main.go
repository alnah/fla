package main

import (
	"os"

	"github.com/alnah/fla/aiclient"
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

	/********* Demo *********/
	tts, err := aiclient.NewTTSClient(
		aiclient.WithProvider[*aiclient.TTSClient](aiclient.ProviderOpenAI),
		aiclient.WithURL[*aiclient.TTSClient](aiclient.URLSpeechAudioOpenAI),
		aiclient.WithAPIKey[*aiclient.TTSClient](aiclient.EnvOpenAIAPIKey),
		aiclient.WithModel[*aiclient.TTSClient](aiclient.AIModelTTSOpenAI),
		aiclient.WithVoice(aiclient.VoiceOpenAIFemaleAlloy),
		aiclient.WithInstructions("test"),
		aiclient.WithText("test"),
	)
	if err != nil {
		log.Error("tts", "error", err.Error())
		return
	}
	byt, err := tts.Do()
	if err != nil {
		log.Error("speech", "error", err.Error())
		return
	}
	file, _ := os.Create("tts.mp3")
	_, _ = file.Write(byt)
	log.Info("done", "audio", file.Name())
}
