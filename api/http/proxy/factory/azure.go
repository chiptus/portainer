package factory

import (
	"net/http"
	"net/url"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/azure"
)

func newAzureProxy(userActivityService portaineree.UserActivityService, endpoint *portaineree.Endpoint, dataStore dataservices.DataStore) (http.Handler, error) {
	remoteURL, err := url.Parse(azureAPIBaseURL)
	if err != nil {
		return nil, err
	}

	proxy := newSingleHostReverseProxyWithHostHeader(remoteURL)
	proxy.Transport = azure.NewTransport(&endpoint.AzureCredentials, userActivityService, dataStore, endpoint)
	return proxy, nil
}
