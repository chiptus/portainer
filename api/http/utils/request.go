package utils

import (
	"bytes"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

// CopyRequestBody copies the request body if it hasn't been read yet
func CopyRequestBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}

	// the implementation is a bit naive as we intend to read the whole body in-memory
	// that might be problematic in case of large payloads, but in a general case shouldn't be a problem
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Debug("failed to read request body")
	}

	r.Body.Close()

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body
}
