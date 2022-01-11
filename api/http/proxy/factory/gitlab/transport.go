package gitlab

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	ru "github.com/portainer/portainer-ee/api/http/utils"
)

type Transport struct {
	httpTransport       *http.Transport
	userActivityService portaineree.UserActivityService
}

// NewTransport returns a pointer to a new instance of Transport that implements the HTTP Transport
// interface for proxying requests to the Gitlab API.
func NewTransport(userActivityService portaineree.UserActivityService) *Transport {
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

	r, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Private-Token", token)

	// need a copy of the request body to preserve the original
	body := ru.CopyRequestBody(r)

	response, err := transport.httpTransport.RoundTrip(r)
	if err == nil {
		useractivity.LogProxiedActivity(transport.userActivityService, nil, response.StatusCode, body, r)
	}

	return response, err
}
