package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/transport"
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

func (t *TTSClient) BaseClient() *baseClient { return t.base }

func WithVoice(v voice) option[*TTSClient] {
	return func(s *TTSClient) { s.voice = v }
}

func WithText(txt Text) option[*TTSClient] {
	return func(s *TTSClient) { s.text = txt }
}
func WithInstructions(i Instructions) option[*TTSClient] {
	return func(s *TTSClient) { s.instructions = i }
}
func WithSpeed(sp Speed) option[*TTSClient] {
	return func(s *TTSClient) { s.speed = sp }
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

func (s TTSClient) MarshalJSON() ([]byte, error) {
	switch {
	case s.useOpenAI:
		type openaiPayload struct {
			Input        string `json:"input"`
			Model        string `json:"model"`
			Voice        string `json:"voice"`
			Instructions string `json:"instructions"`
		}
		payload := openaiPayload{
			Input:        s.text.String(),
			Model:        s.base.model.String(),
			Voice:        s.voice.String(),
			Instructions: s.instructions.String(),
		}
		return json.Marshal(payload)
	case s.useElevenLabs:
		type elevenlabsPayload struct {
			Text          string `json:"text"`
			ModelID       string `json:"model_id"`
			VoiceSettings struct {
				Speed Speed `json:"speed"`
			} `json:"voice_settings"`
		}
		payload := elevenlabsPayload{
			Text:    s.text.String(),
			ModelID: s.base.model.String(),
			VoiceSettings: struct {
				Speed Speed "json:\"speed\""
			}{s.speed},
		}
		return json.Marshal(payload)
	default:
		return nil, errors.New("no provider configured")
	}
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

func (s *TTSClient) applyDefaults() *TTSClient {
	if s.base.ctx == nil {
		s.base.ctx = context.Background()
	}
	if s.base.log == nil {
		s.base.log = logger.New()
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

func (s *TTSClient) validate() error {
	if s.base.ctx == nil {
		return errors.New("context must be provided")
	}
	if s.base.log == nil {
		return errors.New("logger must be set")
	}
	if s.base.httpClient == nil {
		return errors.New("http client must be set")
	}
	if err := s.base.provider.Validate(); err != nil {
		return err
	}
	if err := s.base.url.Validate(); err != nil {
		return err
	}
	if err := s.base.apiKey.Validate(); err != nil {
		return err
	}
	if err := s.base.httpMethod.Validate(); err != nil {
		return err
	}
	if err := s.base.model.Validate(); err != nil {
		return err
	}
	if err := s.voice.Validate(s.base.provider); err != nil {
		return err
	}
	if err := s.text.Validate(); err != nil {
		return err
	}
	if s.useOpenAI == s.useElevenLabs {
		return errors.New("must configure exactly one provider: openai or elevenlabs")
	}
	if s.useOpenAI && s.base.provider != ProviderOpenAI {
		return fmt.Errorf("url indicates openai but provider is %s", s.base.provider)
	}
	if s.useElevenLabs && s.base.provider != ProviderElevenLabs {
		return fmt.Errorf("url indicates elevenlabs but provider is %s", s.base.provider)
	}
	switch {
	case s.useOpenAI:
		if s.base.model != ModelTTSOpenAI {
			return fmt.Errorf("model %s not supported by openai", s.base.model)
		}
		if err := s.instructions.Validate(s.base.provider); err != nil {
			return err
		}
	case s.useElevenLabs:
		if s.base.model != ModelTTSElevenLabs {
			return fmt.Errorf("model %s not supported by elevenlabs", s.base.model)
		}
		if err := s.speed.Validate(s.base.provider); err != nil {
			return err
		}
	}
	return nil
}

func (t *TTSClient) newTransportChain() http.RoundTripper {
	return transport.Chain(
		t.base.httpClient.Transport,
		transport.AddHeader("Content-Type", "application/json"),
		t.addAuthHeader(),
		transport.AddHeader("User-Agent", "Fla/1.0"),
		transport.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, t.buildError()),
		transport.UseCircuitBreaker(breaker.New(breaker.ThirdPartyConfig())),
		transport.UseRetrier(retrier.NewExpBackoffJitter(), isRetryable),
		transport.UseLogger(t.base.log),
	)
}

func (t *TTSClient) addAuthHeader() transport.Middleware {
	if t.base.provider == ProviderOpenAI {
		return transport.AddHeader("Authorization", "Bearer "+t.base.apiKey.GetEnv())
	}
	return transport.AddHeader("xi-api-key", t.base.apiKey.GetEnv())
}

func (t *TTSClient) buildError() transport.ErrorFactoryFunc {
	if t.useOpenAI {
		return buildOpenaiError
	}
	return buildElevenlabsError

}
