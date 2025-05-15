package main

import (
	"os"

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

	/********* Demo *********/
	chat, err := ai.NewChat(
		ai.WithLogger[*ai.Chat](log),
		ai.WithProvider[*ai.Chat](ai.ProviderOpenAI),
		ai.WithURL[*ai.Chat](ai.URLChatCompletionOpenAI),
		ai.WithAPIKey[*ai.Chat](ai.EnvOpenAIAPIKey),
		ai.WithModel[*ai.Chat](ai.AIModelCostOptimizedOpenAI),
		ai.WithMessages(ai.Messages{
			ai.Message{
				Role:    ai.RoleUser,
				Content: "Écrire un texte racontant la routine sportive d'une jeune femme la semaine.",
			},
		}),
	)
	if err != nil {
		log.Error("chat client", "error", err.Error())
		return
	}

	completion, err := chat.Do()
	if err != nil {
		log.Error("chat response", "error", err.Error())
		return
	}
	text := ai.Text(completion.String())
	log.Info("chat response", "completion", text)

	tts, err := ai.NewTTS(
		ai.WithLogger[*ai.TTS](log),
		ai.WithProvider[*ai.TTS](ai.ProviderOpenAI),
		ai.WithURL[*ai.TTS](ai.URLSpeechAudioOpenAI),
		ai.WithAPIKey[*ai.TTS](ai.EnvOpenAIAPIKey),
		ai.WithModel[*ai.TTS](ai.AIModelTTSOpenAI),
		ai.WithVoice(ai.VoiceOpenAIFemaleAlloy),
		ai.WithInstructions(`Personalité : Une jeune femme qui adore le sport, notamment les sports extrêmes.
Ton : Passionnée, enjouée, elle adore partager ses passions.
Prononciation : parle lentement, elle parle à des débutants en français, niveau A1.`),
		ai.WithText(text),
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
