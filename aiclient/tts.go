package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	t.base.httpClient.Transport = t.newTransportChain()
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

func (t *TTSClient) Do() ([]byte, error) {
	byt, err := json.Marshal(t)
	if err != nil {
		return nil, NewTTSClientError(t.base.provider, "failed to marshal payload", err)
	}

	url := t.base.url.String()
	if t.base.provider == ProviderElevenLabs {
		url += "/" + t.Voice.String()
	}
	req, err := http.NewRequestWithContext(t.base.ctx, t.base.httpMethod.String(), url, bytes.NewBuffer(byt))
	if err != nil {
		return nil, NewTTSClientError(t.base.provider, "failed to build http request", err)
	}

	res, err := t.base.httpClient.Do(req)
	if err != nil {
		return nil, NewTTSClientError(t.base.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return nil, NewTTSClientError(t.base.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return nil, BuildProviderError(t.base.provider, res)
	}

	return byt, nil
}
