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

type Speech struct {
	// shared fields
	base *Base
	// api fields
	Voice        voice
	Text         Text
	Instructions Instructions // OpenAI only
	Speed        Speed        // ElevenLabs only
	// provider fields
	useOpenAI     bool
	useElevenLabs bool
}

func NewTTS(options ...option[*Speech]) (*Speech, error) {
	t := &Speech{base: &Base{}}
	for _, opt := range options {
		opt(t)
	}
	if err := t.applyDefaults().setProviderFlag().validate(); err != nil {
		return nil, fmt.Errorf("failed to build speech client: %w", err)
	}
	t.base.httpClient.Transport = t.newTransportChain()
	return t, nil
}

func (s *Speech) applyDefaults() *Speech {
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

func (s *Speech) setProviderFlag() *Speech {
	s.useOpenAI = strings.Contains(s.base.url.String(), ProviderOpenAI.String())
	s.useElevenLabs = strings.Contains(s.base.url.String(), ProviderElevenLabs.String())
	return s
}

func (s *Speech) Do() ([]byte, error) {
	byt, err := json.Marshal(s)
	if err != nil {
		return nil, NewTTSError(s.base.provider, "failed to marshal payload", err)
	}

	url := s.base.url.String()
	if s.base.provider == ProviderElevenLabs {
		url += "/" + s.Voice.String()
	}
	req, err := http.NewRequestWithContext(s.base.ctx, s.base.httpMethod.String(), url, bytes.NewBuffer(byt))
	if err != nil {
		return nil, NewTTSError(s.base.provider, "failed to build http request", err)
	}

	res, err := s.base.httpClient.Do(req)
	if err != nil {
		return nil, NewTTSError(s.base.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return nil, NewTTSError(s.base.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return nil, buildProviderError(s.base.provider, res)
	}

	return byt, nil
}
