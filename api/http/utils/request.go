package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	r.Body.Close()
	if err != nil {
		log.Debug().Err(err).Msg("failed to read request body")
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body
}

// RetrieveArrayQueryParameter returns the value of a query parameter as a string array.
// it will return nil if the query parameter is not found.
//
// Example:
//
//	GET /api/resource?filter=foo&filter=bar
//	RetrieveArrayQueryParameter(request, "filter") => []string{"foo", "bar"}
func RetrieveArrayQueryParameter(r *http.Request, parameter string) []string {
	list, exists := r.Form[fmt.Sprintf("%s[]", parameter)]
	if !exists {
		return nil
	}

	return list
}

func RetrieveNumberArrayQueryParameter[T ~int](r *http.Request, parameter string) ([]T, error) {
	if r.Form == nil {
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("Unable to parse form: %w", err)
		}
	}

	list := RetrieveArrayQueryParameter(r, parameter)
	if list == nil {
		return nil, nil
	}

	var result []T
	for _, item := range list {
		number, err := strconv.Atoi(item)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse parameter %q: %w", parameter, err)

		}

		result = append(result, T(number))
	}

	return result, nil
}
