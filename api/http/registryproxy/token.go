package registryproxy

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
)

type (
	tokenSecuredTransport struct {
		config        *portaineree.RegistryManagementConfiguration
		client        *http.Client
		httpTransport http.RoundTripper
	}

	genericAuthenticationResponse struct {
		AccessToken string `json:"token"`
	}

	azureAuthenticationResponse struct {
		AccessToken string `json:"access_token"`
	}
)

func newTokenSecuredRegistryProxy(uri string, config *portaineree.RegistryManagementConfiguration, httpTransport http.RoundTripper) (http.Handler, error) {
	url, err := url.Parse("https://" + uri)
	if err != nil {
		return nil, err
	}

	proxy := newSingleHostReverseProxyWithHostHeader(url)
	proxy.Transport = &tokenSecuredTransport{
		config:        config,
		httpTransport: httpTransport,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	return proxy, nil
}

// RoundTrip will first send a lightweight copy of the original request (same URL and method) and
// will then inspect the response code of the response.
// If the response code is 401 (Unauthorized), it will send an authentication request
// based on the information retrieved in the Www-Authenticate response header
// (https://docs.docker.com/registry/spec/auth/scope/#resource-provider-use) and
// retrieve an authentication token. It will then retry the original request
// decorated with a new Authorization header containing the authentication token.
func (transport *tokenSecuredTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	requestCopy, err := cloneRequest(request)
	if err != nil {
		return nil, err
	}

	if transport.config.Type == portaineree.EcrRegistry {
		err = registryutils.EnsureManageTokenValid(transport.config)
		if err != nil {
			return nil, err
		}

		requestCopy.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(transport.config.AccessToken)))
	}

	response, err := transport.httpTransport.RoundTrip(requestCopy)
	if err != nil {
		return response, err
	}

	if response.StatusCode == http.StatusUnauthorized {
		token, err := requestToken(response, transport.config)
		if err != nil {
			return response, err
		}

		request.Header.Set("Authorization", "Bearer "+*token)
		response, err = transport.httpTransport.RoundTrip(request)
		if err != nil {
			return nil, err
		}

	}

	return response, nil
}
