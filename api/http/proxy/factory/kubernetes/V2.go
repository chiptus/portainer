package kubernetes

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	useractivityhttp "github.com/portainer/portainer/api/http/useractivity"
	consts "github.com/portainer/portainer/api/useractivity"
)

func (transport *baseTransport) proxyV2Request(request *http.Request, requestPath string) (*http.Response, error) {
	return transport.decorateV2Operation(request)
}

func (transport *baseTransport) decorateV2Operation(request *http.Request) (*http.Response, error) {
	body, err := utils.CopyBody(request)
	if err != nil {
		return nil, err
	}

	response, err := transport.executeKubernetesRequest(request, false)

	if err == nil && (200 <= response.StatusCode && response.StatusCode < 300) {
		transport.logV2Operations(request, body)
	}

	return response, err
}

func (transport *baseTransport) logV2Operations(request *http.Request, body []byte) {
	requestPath := strings.TrimPrefix(request.URL.Path, "/v2")
	var cleanBody interface{}
	var err error
	switch {
	case strings.HasPrefix(requestPath, "/dockerhub"):
		//if request method is POST or PUT
		//make sure the request body is trimmed
		if request.Method == "POST" || request.Method == "PUT" {
			cleanBody, err = hideDockerHubCredentials(body)
			if err != nil {
				log.Printf("[ERROR] [http,dockerhub] [message: failed cleaning request body] [error: %s]", err)
			}
		}
	}
	useractivityhttp.LogHttpActivity(transport.userActivityStore, transport.endpoint.Name, request, cleanBody)
}

// hideDockerHubCredentials removes the confidential properties from the DockerHub payload and returns the new payload
func hideDockerHubCredentials(body []byte) (interface{}, error) {
	payload := &portainer.DockerHub{}
	err := json.Unmarshal(body, payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing body")
	}
	payload.Password = consts.RedactedValue
	return payload, nil
}
