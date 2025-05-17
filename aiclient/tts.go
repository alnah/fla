package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alnah/fla/logger"
)

type TTSClient struct {
	// shared fields
	base *baseClient
	// api fields
	voice        voice
	text         Text
	instructions Instructions // OpenAI only
	speed        Speed        // ElevenLabs only
	// provider fields
	useOpenAI     bool
	useElevenLabs bool
}

func NewTTSClient(options ...option[*TTSClient]) (*TTSClient, error) {
	s := &TTSClient{base: &baseClient{}}
	for _, opt := range options {
		opt(s)
	}
	if err := s.applyDefaults().setProviderFlag().validate(); err != nil {
		return nil, NewTTSClientError(s.base.provider, "failed to build text-to-speech client", err)
	}
	return s, nil
}

func (s *TTSClient) applyDefaults() *TTSClient {
	if s.base.ctx == nil {
		s.base.ctx = context.Background()
	}
	if s.base.logger == nil {
		s.base.logger = logger.New()
	}
	if s.base.httpClient == nil {
		s.base.httpClient = &http.Client{Timeout: 10 * time.Minute}
	}
	if s.base.httpMethod == "" {
		s.base.httpMethod = httpMethod(http.MethodPost)
	}
	return s
}

func (s *TTSClient) setProviderFlag() *TTSClient {
	s.useOpenAI = strings.Contains(s.base.url.String(), ProviderOpenAI.String())
	s.useElevenLabs = strings.Contains(s.base.url.String(), ProviderElevenLabs.String())
	return s
}

func (s *TTSClient) Audio() ([]byte, error) {
	byt, err := json.Marshal(s)
	if err != nil {
		return nil, NewTTSClientError(s.base.provider, "failed to marshal payload", err)
	}

	url := s.base.url.String()
	if s.base.provider == ProviderElevenLabs {
		url += "/" + s.voice.String()
	}
	req, err := http.NewRequestWithContext(s.base.ctx, s.base.httpMethod.String(), url, bytes.NewBuffer(byt))
	if err != nil {
		return nil, NewTTSClientError(s.base.provider, "failed to build http request", err)
	}

	s.base.httpClient.Transport = s.newTransportChain()
	res, err := s.base.httpClient.Do(req)
	if err != nil {
		return nil, NewTTSClientError(s.base.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return nil, NewTTSClientError(s.base.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return nil, buildProviderError(s.base.provider, res)
	}

	return byt, nil
}
