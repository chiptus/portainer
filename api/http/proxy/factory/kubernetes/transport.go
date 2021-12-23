package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	ru "github.com/portainer/portainer-ee/api/http/utils"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/pkg/errors"
)

type baseTransport struct {
	httpTransport       *http.Transport
	tokenManager        *tokenManager
	endpoint            *portaineree.Endpoint
	userActivityService portaineree.UserActivityService
	k8sClientFactory    *cli.ClientFactory
	dataStore           portaineree.DataStore
}

func newBaseTransport(httpTransport *http.Transport, tokenManager *tokenManager, endpoint *portaineree.Endpoint, userActivityService portaineree.UserActivityService, k8sClientFactory *cli.ClientFactory, dataStore portaineree.DataStore) *baseTransport {
	return &baseTransport{
		httpTransport:       httpTransport,
		tokenManager:        tokenManager,
		endpoint:            endpoint,
		userActivityService: userActivityService,
		k8sClientFactory:    k8sClientFactory,
		dataStore:           dataStore,
	}
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *baseTransport) prepareRoundTrip(request *http.Request) error {
	token, err := getRoundTripToken(request, transport.tokenManager, transport.endpoint.ID)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return nil
}

// proxyKubernetesRequest intercepts a Kubernetes API request and apply logic based
// on the requested operation.
func (transport *baseTransport) proxyKubernetesRequest(request *http.Request) (*http.Response, error) {
	// URL path examples:
	// http://localhost:9000/api/endpoints/3/kubernetes/api/v1/namespaces
	// http://localhost:9000/api/endpoints/3/kubernetes/apis/apps/v1/namespaces/default/deployments
	apiVersionRe := regexp.MustCompile(`^(/kubernetes)?/(api|apis/apps)/v[0-9](\.[0-9])?`)
	requestPath := apiVersionRe.ReplaceAllString(request.URL.Path, "")

	switch {
	case strings.EqualFold(requestPath, "/namespaces"):
		return transport.executeKubernetesRequest(request)
	case strings.HasPrefix(requestPath, "/namespaces"):
		return transport.proxyNamespacedRequest(request, requestPath)
	default:
		return transport.executeKubernetesRequest(request)
	}
}

func (transport *baseTransport) proxyNamespacedRequest(request *http.Request, fullRequestPath string) (*http.Response, error) {
	requestPath := strings.TrimPrefix(fullRequestPath, "/namespaces/")
	split := strings.SplitN(requestPath, "/", 2)
	namespace := split[0]

	requestPath = ""
	if len(split) > 1 {
		requestPath = split[1]
	}

	switch {
	case strings.HasPrefix(requestPath, "pods"):
		return transport.proxyPodsRequest(request, namespace, requestPath)
	case strings.HasPrefix(requestPath, "deployments"):
		return transport.proxyDeploymentsRequest(request, namespace, requestPath)
	case requestPath == "" && request.Method == "DELETE":
		return transport.proxyNamespaceDeleteOperation(request, namespace)
	default:
		return transport.executeKubernetesRequest(request)
	}
}

func (transport *baseTransport) executeKubernetesRequest(request *http.Request) (*http.Response, error) {
	// need a copy of the request body to preserve the original
	body := ru.CopyRequestBody(request)

	resp, err := transport.httpTransport.RoundTrip(request)
	if err == nil {
		useractivity.LogProxiedActivity(transport.userActivityService, transport.endpoint, resp.StatusCode, body, request)
	}

	// This fix was made to resolve a k8s e2e test, more detailed investigation should be done later.
	if err == nil && resp.StatusCode == http.StatusMovedPermanently {
		oldLocation := resp.Header.Get("Location")
		if oldLocation != "" {
			stripedPrefix := strings.TrimSuffix(request.RequestURI, request.URL.Path)
			// local proxy strips "/kubernetes" but agent proxy and edge agent proxy do not
			stripedPrefix = strings.TrimSuffix(stripedPrefix, "/kubernetes")
			newLocation := stripedPrefix + "/kubernetes" + oldLocation
			resp.Header.Set("Location", newLocation)
		}
	}

	return resp, err
}

func (transport *baseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport.proxyKubernetesRequest(request)
}

func getRoundTripToken(request *http.Request, tokenManager *tokenManager, endpointID portaineree.EndpointID) (string, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return "", err
	}

	var token string
	if tokenData.Role == portaineree.AdministratorRole {
		token = tokenManager.GetAdminServiceAccountToken()
	} else {
		token, err = tokenManager.GetUserServiceAccountToken(int(tokenData.ID), int(endpointID))
		if err != nil {
			log.Printf("Failed retrieving service account token: %v", err)
			return "", err
		}
	}

	return token, nil
}

func decorateAgentRequest(r *http.Request, dataStore portaineree.DataStore) error {
	requestPath := strings.TrimPrefix(r.URL.Path, "/v2")

	switch {
	case strings.HasPrefix(requestPath, "/dockerhub"):
		decorateAgentDockerHubRequest(r, dataStore)
	}

	return nil
}

func decorateAgentDockerHubRequest(r *http.Request, dataStore portaineree.DataStore) error {
	requestPath, registryIdString := path.Split(r.URL.Path)

	registryID, err := strconv.Atoi(registryIdString)
	if err != nil {
		return fmt.Errorf("missing registry id: %w", err)
	}

	r.URL.Path = strings.TrimSuffix(requestPath, "/")

	registry := &portaineree.Registry{
		Type: portaineree.DockerHubRegistry,
	}

	if registryID != 0 {
		registry, err = dataStore.Registry().Registry(portaineree.RegistryID(registryID))
		if err != nil {
			return fmt.Errorf("failed fetching registry: %w", err)
		}
	}

	if registry.Type != portaineree.DockerHubRegistry {
		return errors.New("invalid registry type")
	}

	newBody, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	r.Method = http.MethodPost

	r.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	r.ContentLength = int64(len(newBody))

	return nil
}
