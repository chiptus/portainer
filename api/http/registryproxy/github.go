package registryproxy

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"

	portaineree "github.com/portainer/portainer-ee/api"
)

type githubTransport struct {
	config        *portaineree.RegistryManagementConfiguration
	httpTransport http.RoundTripper
}

func newGithubRegistryProxy(uri string, config *portaineree.RegistryManagementConfiguration, httpTransport http.RoundTripper) (http.Handler, error) {
	scheme := "https"
	url, err := url.Parse(scheme + "://" + uri)
	if err != nil {
		return nil, err
	}

	url.Scheme = scheme

	proxy := newSingleHostReverseProxyWithHostHeader(url)
	proxy.Transport = &githubTransport{
		config:        config,
		httpTransport: httpTransport,
	}

	return proxy, nil
}

// RoundTrip will simply check if the configuration associated to the
// custom registry has a token saved in it and add it in the request
// to authenticate on the github API.
func (transport *githubTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	token := transport.config.Password
	if token == "" {
		return nil, errors.New("No github token provided")
	}

	request.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(token)))

	response, err := transport.httpTransport.RoundTrip(request)

	return response, err
}
