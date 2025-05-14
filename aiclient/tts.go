package aiclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alnah/fla/logger"
)

type TTSClient struct {
	// http fields
	base *BaseClient
	// api fields
	Voice        Voice
	Text         Text
	Instructions Instructions // OpenAI only
	Speed        Speed        // ElevenLabs only
	// provider fields
	UseOpenAI     bool
	UseElevenLabs bool
}

func NewTTSClient(options ...Option[*TTSClient]) (*TTSClient, error) {
	t := &TTSClient{base: &BaseClient{}}
	for _, opt := range options {
		opt(t)
	}
	if err := t.applyDefaults().setProviderFlag().validate(); err != nil {
		return nil, fmt.Errorf("failed to build tts client: %w", err)
	}
	return t, nil
}

func (t *TTSClient) applyDefaults() *TTSClient {
	if t.base.ctx == nil {
		t.base.ctx = context.Background()
	}
	if t.base.logger == nil {
		t.base.logger = logger.New()
	}
	if t.base.httpClient == nil {
		t.base.httpClient = &http.Client{Timeout: 10 * time.Minute}
	}
	if t.base.httpMethod == "" {
		t.base.httpMethod = HTTPMethod(http.MethodPost)
	}
	return t
}

func (t *TTSClient) setProviderFlag() *TTSClient {
	t.UseOpenAI = strings.Contains(t.base.url.String(), ProviderOpenAI.String())
	t.UseElevenLabs = strings.Contains(t.base.url.String(), ProviderElevenLabs.String())
	return t
}
