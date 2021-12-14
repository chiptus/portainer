package gitlab

import (
	"errors"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/useractivity"
	ru "github.com/portainer/portainer/api/http/utils"
)

type Transport struct {
	httpTransport       *http.Transport
	userActivityService portainer.UserActivityService
}

// NewTransport returns a pointer to a new instance of Transport that implements the HTTP Transport
// interface for proxying requests to the Gitlab API.
func NewTransport(userActivityService portainer.UserActivityService) *Transport {
	return &Transport{
		userActivityService: userActivityService,
		httpTransport:       &http.Transport{},
	}
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	token := request.Header.Get("Private-Token")
	if token == "" {
		return nil, errors.New("no gitlab token provided")
	}

	r, err := http.NewRequest(request.Method, request.URL.String(), nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Private-Token", token)

	// need a copy of the request body to preserve the original
	body := ru.CopyRequestBody(request)

	response, err := transport.httpTransport.RoundTrip(request)
	if err == nil {
		useractivity.LogProxiedActivity(transport.userActivityService, nil, response.StatusCode, body, request)
	}

	return response, err
}
