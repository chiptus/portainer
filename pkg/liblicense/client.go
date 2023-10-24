package liblicense

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultHTTPTimeout = 5
)

// Post executes a simple HTTP POST to the specified URL with data specified as payload and returns
// the content of the response body. Timeout can be specified via the timeout parameter,
// will default to defaultHTTPTimeout if set to 0.
func Post(url string, data []byte, timeout int) ([]byte, error) {
	if timeout == 0 {
		timeout = defaultHTTPTimeout
	}

	client := &http.Client{
		Timeout: time.Second * time.Duration(timeout),
	}

	response, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode/100 != 2 {
		return nil, fmt.Errorf("Invalid response status: %d, body: %s", response.StatusCode, body)
	}

	return body, nil
}
