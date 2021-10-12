package kubernetes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/kubernetes/cli"
	"github.com/sirupsen/logrus"

	portainer "github.com/portainer/portainer/api"
)

type baseTransport struct {
	httpTransport     *http.Transport
	tokenManager      *tokenManager
	endpoint          *portainer.Endpoint
	userActivityStore portainer.UserActivityStore
	k8sClientFactory  *cli.ClientFactory
	dataStore         portainer.DataStore
}

func newBaseTransport(httpTransport *http.Transport, tokenManager *tokenManager, endpoint *portainer.Endpoint, userActivityStore portainer.UserActivityStore, k8sClientFactory *cli.ClientFactory, dataStore portainer.DataStore) *baseTransport {
	return &baseTransport{
		httpTransport:     httpTransport,
		tokenManager:      tokenManager,
		endpoint:          endpoint,
		userActivityStore: userActivityStore,
		k8sClientFactory:  k8sClientFactory,
		dataStore:         dataStore,
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
	apiVersionRe := regexp.MustCompile(`^(/kubernetes)?/api/v[0-9](\.[0-9])?`)
	requestPath := apiVersionRe.ReplaceAllString(request.URL.Path, "")

	switch {
	case strings.EqualFold(requestPath, "/namespaces"):
		return transport.executeKubernetesRequest(request, true)
	case strings.HasPrefix(requestPath, "/namespaces"):
		return transport.proxyNamespacedRequest(request, requestPath)
	case strings.HasPrefix(requestPath, "/v2"):
		return transport.proxyV2Request(request, requestPath)
	default:
		return transport.executeKubernetesRequest(request, true)
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
	case strings.HasPrefix(requestPath, "configmaps"):
		return transport.proxyConfigMapsRequest(request, requestPath)
	case strings.HasPrefix(requestPath, "secrets"):
		return transport.proxySecretsRequest(request, namespace, requestPath)
	case requestPath == "" && request.Method == "DELETE":
		return transport.proxyNamespaceDeleteOperation(request, namespace)
	default:
		return transport.executeKubernetesRequest(request, true)
	}
}

func (transport *baseTransport) executeKubernetesRequest(request *http.Request, shouldLog bool) (*http.Response, error) {
	var body []byte

	if shouldLog {
		bodyBytes, err := utils.CopyBody(request)
		if err != nil {
			logrus.WithError(err).Debug("[k8s transport] failed parsing body")
		}

		body = bodyBytes
	}

	resp, err := transport.httpTransport.RoundTrip(request)

	// log if request is success
	if shouldLog && err == nil && (200 <= resp.StatusCode && resp.StatusCode < 300) {
		useractivity.LogProxyActivity(transport.userActivityStore, transport.endpoint.Name, request, body)
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

func getRoundTripToken(request *http.Request, tokenManager *tokenManager, endpointID portainer.EndpointID) (string, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return "", err
	}

	var token string
	if tokenData.Role == portainer.AdministratorRole {
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

func decorateAgentRequest(r *http.Request, dataStore portainer.DataStore) error {
	requestPath := strings.TrimPrefix(r.URL.Path, "/v2")

	switch {
	case strings.HasPrefix(requestPath, "/dockerhub"):
		decorateAgentDockerHubRequest(r, dataStore)
	}

	return nil
}

func decorateAgentDockerHubRequest(r *http.Request, dataStore portainer.DataStore) error {
	requestPath, registryIdString := path.Split(r.URL.Path)

	registryID, err := strconv.Atoi(registryIdString)
	if err != nil {
		return fmt.Errorf("missing registry id: %w", err)
	}

	r.URL.Path = strings.TrimSuffix(requestPath, "/")

	registry := &portainer.Registry{
		Type: portainer.DockerHubRegistry,
	}

	if registryID != 0 {
		registry, err = dataStore.Registry().Registry(portainer.RegistryID(registryID))
		if err != nil {
			return fmt.Errorf("failed fetching registry: %w", err)
		}
	}

	if registry.Type != portainer.DockerHubRegistry {
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
