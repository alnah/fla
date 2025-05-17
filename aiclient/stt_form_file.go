package aiclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
)

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
