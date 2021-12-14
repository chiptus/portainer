package utils

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CopyRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		expected []byte
	}{
		{
			name:     "empty payload",
			payload:  nil,
			expected: []byte{},
		},
		{
			name:     "non-empty payload",
			payload:  []byte("payload"),
			expected: []byte("payload"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.payload))

			copiedBody := CopyRequestBody(r)
			require.Equal(t, tt.expected, copiedBody)

			requestBody, _ := io.ReadAll(r.Body)
			r.Body.Close()
			require.Equal(t, tt.expected, requestBody)
		})
	}

}
