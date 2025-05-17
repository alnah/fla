package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	fu "github.com/alnah/fla/fileutil"
	"github.com/alnah/fla/logger"
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
