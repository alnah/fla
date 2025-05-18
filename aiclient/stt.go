package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alnah/fla/breaker"
	fu "github.com/alnah/fla/fileutil"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/transport"
)

type STTClient struct {
	// shared fields
	base *baseClient
	// api fields
	file     *os.File
	filePath fu.FilePath
	language ISO6391
	// provider fields
	useOpenAI     bool
	useElevenLabs bool
	// internal
	filePathSecure string
	contentType    string
	formFileBody   *bytes.Buffer
}

func (t *STTClient) BaseClient() *baseClient { return t.base }

func WithFilePath(f fu.FilePath) option[*STTClient] {
	return func(t *STTClient) { t.filePath = f }
}

func WithLanguage(i ISO6391) option[*STTClient] {
	return func(t *STTClient) { t.language = i }
}

func NewSTTClient(options ...option[*STTClient]) (*STTClient, error) {
	t := &STTClient{base: &baseClient{}}
	for _, opt := range options {
		opt(t)
	}
	var err error
	t.filePathSecure, err = t.applyDefaults().setProviderFlag().validate()
	if err != nil {
		return nil, NewSTTClientError(t.base.provider, "failed to build speech-to-text client", err)
	}
	t.file, err = os.Open(t.filePathSecure)
	if err != nil {
		return nil, NewSTTClientError(t.base.provider, "failed to build speech-to-text client", err)
	}
	return t, nil
}

func (s *STTClient) Transcript() (transcriptResponse, error) {
	err := s.newFormFileBody()
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(s.base.provider, "failed to build form file body", err)
	}

	req, err := http.NewRequestWithContext(s.base.ctx, s.base.httpMethod.String(), s.base.url.String(), s.formFileBody)
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(s.base.provider, "failed to build http request", err)
	}

	s.base.httpClient.Transport = s.newTransportChain()
	res, err := s.base.httpClient.Do(req)
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(s.base.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err := io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(s.base.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return transcriptResponse{}, buildProviderError(s.base.provider, res)
	}

	var transcript transcriptResponse
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&transcript); err != nil {
		return transcriptResponse{}, NewSTTClientError(s.base.provider, "failed to decode response body", err)
	}
	return transcript, nil
}

func (s *STTClient) applyDefaults() *STTClient {
	if s.base.ctx == nil {
		s.base.ctx = context.Background()
	}
	if s.base.log == nil {
		s.base.log = logger.New()
	}
	if s.base.httpClient == nil {
		s.base.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if s.base.httpMethod == "" {
		s.base.httpMethod = httpMethod(http.MethodPost)
	}
	return s
}

func (s *STTClient) setProviderFlag() *STTClient {
	s.useOpenAI = strings.Contains(s.base.url.String(), ProviderOpenAI.String())
	s.useElevenLabs = strings.Contains(s.base.url.String(), ProviderElevenLabs.String())
	return s
}

func (s *STTClient) validate() (string, error) {
	if s.base.ctx == nil {
		return "", errors.New("context must be provided")
	}
	if s.base.log == nil {
		return "", errors.New("logger must be set")
	}
	if s.base.httpClient == nil {
		return "", errors.New("http client must be set")
	}
	if err := s.base.provider.Validate(); err != nil {
		return "", err
	}
	if err := s.base.url.Validate(); err != nil {
		return "", err
	}
	if err := s.base.apiKey.Validate(); err != nil {
		return "", err
	}
	if err := s.base.httpMethod.Validate(); err != nil {
		return "", err
	}
	if err := s.base.model.Validate(); err != nil {
		return "", err
	}
	if err := s.language.Validate(); err != nil {
		return "", err
	}
	if s.useOpenAI == s.useElevenLabs {
		return "", errors.New("must configure exactly one provider: openai or elevenlabs")
	}
	if s.useOpenAI && s.base.provider != ProviderOpenAI {
		return "", fmt.Errorf("url indicates openai but provider is %s", s.base.provider)
	}
	if s.useElevenLabs && s.base.provider != ProviderElevenLabs {
		return "", fmt.Errorf("url indicates elevenlabs but provider is %s", s.base.provider)
	}
	allowed := []string{"flac", "mp3", "mp4", "mpeg", "mpga", "m4a", "ogg", "wav", "webm"}
	switch {
	case s.useOpenAI:
		if s.base.model != ModelSTTOpenAI {
			return "", fmt.Errorf("model %s not supported by openai", s.base.model)
		}
		return s.filePath.Validate(25, allowed...)
	case s.useElevenLabs:
		if s.base.model != ModelSTTElevenLabs {
			return "", fmt.Errorf("model %s not supported by elevenlabs", s.base.model)
		}
		return s.filePath.Validate(25, allowed...)
	}
	return "", nil
}

func (s *STTClient) newTransportChain() http.RoundTripper {
	return transport.Chain(
		s.base.httpClient.Transport,
		transport.AddHeader("Content-Type", s.contentType),
		s.addAuthHeader(),
		transport.AddHeader("User-Agent", "Fla/1.0"),
		transport.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, s.buildError()),
		transport.UseCircuitBreaker(breaker.New(breaker.ThirdPartyConfig())),
		transport.UseRetrier(retrier.NewExpBackoffJitter(), isRetryable),
		transport.UseLogger(s.base.log),
	)
}

func (s *STTClient) addAuthHeader() transport.Middleware {
	if s.base.provider == ProviderOpenAI {
		return transport.AddHeader("Authorization", "Bearer "+s.base.apiKey.GetEnv())
	}
	return transport.AddHeader("xi-api-key", s.base.apiKey.GetEnv())

}

func (s *STTClient) buildError() transport.ErrorFactoryFunc {
	if s.useOpenAI {
		return buildOpenaiError
	}
	return buildElevenlabsError
}

func (s *STTClient) newFormFileBody() error {
	s.formFileBody = &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(s.formFileBody)

	// add file field
	part, err := multipartWriter.CreateFormFile("file", filepath.Base(s.filePathSecure))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, s.file); err != nil {
		return fmt.Errorf("failed to copy file to form: %w", err)
	}

	// add text fields
	if s.base.provider == ProviderOpenAI {
		_ = multipartWriter.WriteField("model", s.base.model.String())
	}
	if s.base.provider == ProviderElevenLabs {
		_ = multipartWriter.WriteField("model_id", s.base.model.String())
	}
	_ = multipartWriter.WriteField("language", s.language.String())
	s.contentType = multipartWriter.FormDataContentType()

	// finalize multipart body
	if err = multipartWriter.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return nil
}

// transcriptResponse holds the text after the transcription is done.
type transcriptResponse struct {
	Text string `json:"text,omitempty"`
}

// Content returns the text content from a transcription.
func (t transcriptResponse) Content() string { return t.Text }
