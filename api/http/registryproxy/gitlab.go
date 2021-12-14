package registryproxy

import (
	"errors"
	"net/http"
	"net/url"

	portainer "github.com/portainer/portainer/api"
)

type gitlabTransport struct {
	config              *portainer.RegistryManagementConfiguration
	httpTransport       http.RoundTripper
	userActivityService portainer.UserActivityService
}

func newGitlabRegistryProxy(uri string, config *portainer.RegistryManagementConfiguration, httpTransport http.RoundTripper) (http.Handler, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	proxy := newSingleHostReverseProxyWithHostHeader(url)
	proxy.Transport = &gitlabTransport{
		config:        config,
		httpTransport: httpTransport,
	}

	return proxy, nil
}

// RoundTrip will simply check if the configuration associated to the
// custom registry has a token saved in it and add it in the request
// to authenticate on the gitlab API.
// As Gitlab API also supports Bearer token a new request is created to avoid transmitting
// Portainer Bearer Token, that will conflict with the Private-Token
func (transport *gitlabTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	token := transport.config.Password
	if token == "" {
		return nil, errors.New("No gitlab token provided")
	}

	r, err := http.NewRequest(request.Method, request.URL.String(), nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Private-Token", token)

	response, err := transport.httpTransport.RoundTrip(request)

	return response, err
}
