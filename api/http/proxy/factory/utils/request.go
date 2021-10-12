package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

// CopyBody copies the request body and recreates it
func CopyBody(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}

	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read body")
	}

	request.Body.Close()
	// recreate body to pass to actual request handler
	request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

// GetRequestAsMap returns the response content as a generic JSON object
func GetRequestAsMap(request *http.Request) (map[string]interface{}, error) {
	data, err := getRequestBody(request)
	if err != nil {
		return nil, err
	}

	o, ok := data.(map[string]interface{})
	if !ok {
		return nil, nil
	}
	return o, nil
}

// RewriteRequest will replace the existing request body with the one specified
// in parameters
func RewriteRequest(request *http.Request, newData interface{}) error {
	data, err := marshal(getContentType(request.Header), newData)
	if err != nil {
		return err
	}

	body := ioutil.NopCloser(bytes.NewReader(data))

	request.Body = body
	request.ContentLength = int64(len(data))

	if request.Header == nil {
		request.Header = make(http.Header)
	}
	request.Header.Set("Content-Length", strconv.Itoa(len(data)))

	return nil
}

func getRequestBody(request *http.Request) (interface{}, error) {
	isGzip := request.Header.Get("Content-Encoding") == "gzip"

	return getBody(request.Body, getContentType(request.Header), isGzip)
}
