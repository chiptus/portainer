package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// GetResponseAsJSONObject returns the response content as a generic JSON object
func GetResponseAsJSONObject(response *http.Response) (map[string]interface{}, error) {
	responseData, err := getResponseBody(response)
	if err != nil {
		return nil, err
	}

	responseObject, ok := responseData.(map[string]interface{})
	if !ok {
		return nil, nil
	}
	return responseObject, nil
}

// GetResponseAsJSONArray returns the response content as an array of generic JSON object
func GetResponseAsJSONArray(response *http.Response) ([]interface{}, error) {
	responseData, err := getResponseBody(response)
	if err != nil {
		return nil, err
	}
	if responseData == nil {
		return nil, nil
	}

	switch responseObject := responseData.(type) {
	case []interface{}:
		return responseObject, nil
	case map[string]interface{}:
		if responseObject["message"] != nil {
			return nil, errors.New(responseObject["message"].(string))
		}

		log.Error().
			Str("response", fmt.Sprintf("%+v", responseObject)).
			Msg("invalid response format, expecting JSON array")

		return nil, errors.New("unable to parse response: expected JSON array, got JSON object")
	default:
		log.Error().
			Str("response", fmt.Sprintf("%+v", responseObject)).
			Msg("invalid response format, expecting JSON array")

		return nil, errors.New("unable to parse response: expected JSON array")
	}
}

type errorResponse struct {
	Message string `json:"message,omitempty"`
}

// WriteAccessDeniedResponse will create a new access denied response
func WriteAccessDeniedResponse() (*http.Response, error) {
	response := &http.Response{}
	err := RewriteResponse(response, errorResponse{Message: "access denied to resource"}, http.StatusForbidden)

	return response, err
}

// RewriteAccessDeniedResponse will overwrite the existing response with an access denied response
func RewriteAccessDeniedResponse(response *http.Response) error {
	return RewriteResponse(response, errorResponse{Message: "access denied to resource"}, http.StatusForbidden)
}

// RewriteResponse will replace the existing response body and status code with the one specified
// in parameters
func RewriteResponse(response *http.Response, newResponseData interface{}, statusCode int) error {
	data, err := marshal(getContentType(response.Header), newResponseData)
	if err != nil {
		return err
	}

	body := io.NopCloser(bytes.NewReader(data))

	response.StatusCode = statusCode
	response.Body = body
	response.ContentLength = int64(len(data))

	if response.Header == nil {
		response.Header = make(http.Header)
	}
	response.Header.Set("Content-Length", strconv.Itoa(len(data)))

	return nil
}

func getResponseBody(response *http.Response) (interface{}, error) {
	isGzip := response.Header.Get("Content-Encoding") == "gzip"

	if isGzip {
		response.Header.Del("Content-Encoding")
	}

	bodyBytes, err := CopyResponseBody(response)
	if err != nil {
		return nil, err
	}

	return getBody(io.NopCloser(bytes.NewBuffer(bodyBytes)), getContentType(response.Header), isGzip)
}

// CopyBody copies the request body and recreates it
func CopyResponseBody(response *http.Response) ([]byte, error) {
	if response.Body == nil {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read body")
	}

	response.Body.Close()
	// recreate body to pass to actual request handler
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

func getContentType(headers http.Header) string {
	return headers.Get("Content-type")
}
