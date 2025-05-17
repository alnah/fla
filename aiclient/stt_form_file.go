package aiclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
)

func (t *STTClient) newFormFileBody() error {
	t.formFileBody = &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(t.formFileBody)

	// add file field
	part, err := multipartWriter.CreateFormFile("file", filepath.Base(t.filePathSecure))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, t.file); err != nil {
		return fmt.Errorf("failed to copy file to form: %w", err)
	}

	// add text fields
	if t.base.provider == ProviderOpenAI {
		_ = multipartWriter.WriteField("model", t.base.model.String())
	}
	if t.base.provider == ProviderElevenLabs {
		_ = multipartWriter.WriteField("model_id", t.base.model.String())
	}
	_ = multipartWriter.WriteField("language", t.language.String())
	t.contentType = multipartWriter.FormDataContentType()

	// finalize multipart body
	if err = multipartWriter.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return nil
}
