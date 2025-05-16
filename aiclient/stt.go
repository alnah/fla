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
		return nil, NewSTTClientError(t.base.provider, "failed to build transcript client", err)
	}
	t.file, err = os.Open(t.filePathSecure)
	if err != nil {
		return nil, NewSTTClientError(t.base.provider, "failed to build transcript client", err)
	}
	return t, nil
}

func (t *STTClient) applyDefaults() *STTClient {
	if t.base.ctx == nil {
		t.base.ctx = context.Background()
	}
	if t.base.logger == nil {
		t.base.logger = logger.New()
	}
	if t.base.httpClient == nil {
		t.base.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if t.base.httpMethod == "" {
		t.base.httpMethod = httpMethod(http.MethodPost)
	}
	return t
}

func (t *STTClient) setProviderFlag() *STTClient {
	t.useOpenAI = strings.Contains(t.base.url.String(), ProviderOpenAI.String())
	t.useElevenLabs = strings.Contains(t.base.url.String(), ProviderElevenLabs.String())
	return t
}

func (t *STTClient) Transcript() (transcriptResponse, error) {
	err := t.newFormFileBody()
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(t.base.provider, "failed to build form file body", err)
	}

	req, err := http.NewRequestWithContext(t.base.ctx, t.base.httpMethod.String(), t.base.url.String(), t.formFileBody)
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(t.base.provider, "failed to build http request", err)
	}

	t.base.httpClient.Transport = t.newTransportChain()
	res, err := t.base.httpClient.Do(req)
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(t.base.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err := io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return transcriptResponse{}, NewSTTClientError(t.base.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return transcriptResponse{}, buildProviderError(t.base.provider, res)
	}

	var transcript transcriptResponse
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&transcript); err != nil {
		return transcriptResponse{}, NewSTTClientError(t.base.provider, "failed to decode response body", err)
	}
	return transcript, nil
}
