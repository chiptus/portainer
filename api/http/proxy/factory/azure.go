package factory

import (
	"net/http"
	"net/url"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/azure"
)

func newAzureProxy(userActivityService portainer.UserActivityService, endpoint *portainer.Endpoint, dataStore portainer.DataStore) (http.Handler, error) {
	remoteURL, err := url.Parse(azureAPIBaseURL)
	if err != nil {
		return nil, err
	}

	proxy := newSingleHostReverseProxyWithHostHeader(remoteURL)
	proxy.Transport = azure.NewTransport(&endpoint.AzureCredentials, userActivityService, dataStore, endpoint)
	return proxy, nil
}
