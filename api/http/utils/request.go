package utils

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// CopyRequestBody copies the request body if it hasn't been read yet
func CopyRequestBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}

	//for upload images and files, skip log the body
	if (strings.Contains(r.URL.Path, "browse/put") && r.Method == "POST") ||
		(strings.Contains(r.URL.Path, "images/load") && r.Method == "POST") {
		return nil
	}

	// the implementation is a bit naive as we intend to read the whole body in-memory
	// that might be problematic in case of large payloads, but in a general case shouldn't be a problem
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Debug().Err(err).Msg("failed to read request body")
	}

	r.Body.Close()

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body
}
