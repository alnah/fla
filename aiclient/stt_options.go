package aiclient

import fu "github.com/alnah/fla/fileutil"

func (t *STTClient) BaseClient() *baseClient { return t.base }

func WithFilePath(f fu.FilePath) option[*STTClient] {
	return func(t *STTClient) { t.filePath = f }
}

func WithLanguage(i ISO6391) option[*STTClient] {
	return func(t *STTClient) { t.language = i }
}
