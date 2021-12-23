package azure

import (
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	ru "github.com/portainer/portainer-ee/api/http/utils"
)

type (
	azureAPIToken struct {
		value          string
		expirationTime time.Time
	}

	Transport struct {
		credentials         *portaineree.AzureCredentials
		client              *client.HTTPClient
		token               *azureAPIToken
		mutex               sync.Mutex
		dataStore           portaineree.DataStore
		endpoint            *portaineree.Endpoint
		userActivityService portaineree.UserActivityService
	}

	azureRequestContext struct {
		isAdmin                bool
		endpointResourceAccess bool
		userID                 portaineree.UserID
		userTeamIDs            []portaineree.TeamID
		resourceControls       []portaineree.ResourceControl
	}
)

// NewTransport returns a pointer to a new instance of Transport that implements the HTTP Transport
// interface for proxying requests to the Azure API.
func NewTransport(credentials *portaineree.AzureCredentials, userActivityService portaineree.UserActivityService, dataStore portaineree.DataStore, endpoint *portaineree.Endpoint) *Transport {
	return &Transport{
		credentials:         credentials,
		client:              client.NewHTTPClient(),
		dataStore:           dataStore,
		endpoint:            endpoint,
		userActivityService: userActivityService,
	}
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport.proxyAzureRequest(request)
}

func (transport *Transport) proxyAzureRequest(request *http.Request) (*http.Response, error) {
	requestPath := request.URL.Path

	err := transport.retrieveAuthenticationToken()
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+transport.token.value)

	if match, _ := path.Match(portaineree.AzurePathContainerGroups, requestPath); match {
		return transport.proxyContainerGroupsRequest(request)
	} else if match, _ := path.Match(portaineree.AzurePathContainerGroup, requestPath); match {
		return transport.proxyContainerGroupRequest(request)
	}

	// need a copy of the request body to preserve the original
	body := ru.CopyRequestBody(request)

	response, err := http.DefaultTransport.RoundTrip(request)
	if err == nil {
		useractivity.LogProxiedActivity(transport.userActivityService, transport.endpoint, response.StatusCode, body, request)
	}

	return response, err
}

func (transport *Transport) authenticate() error {
	token, err := transport.client.ExecuteAzureAuthenticationRequest(transport.credentials)
	if err != nil {
		return err
	}

	expiresOn, err := strconv.ParseInt(token.ExpiresOn, 10, 64)
	if err != nil {
		return err
	}

	transport.token = &azureAPIToken{
		value:          token.AccessToken,
		expirationTime: time.Unix(expiresOn, 0),
	}

	return nil
}

func (transport *Transport) retrieveAuthenticationToken() error {
	transport.mutex.Lock()
	defer transport.mutex.Unlock()

	if transport.token == nil {
		return transport.authenticate()
	}

	timeLimit := time.Now().Add(-5 * time.Minute)
	if timeLimit.After(transport.token.expirationTime) {
		return transport.authenticate()
	}

	return nil
}
