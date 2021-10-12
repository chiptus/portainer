package kubernetes

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	useractivityhttp "github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/useractivity"
	"github.com/sirupsen/logrus"
)

func (transport *baseTransport) proxyConfigMapsRequest(request *http.Request, requestPath string) (*http.Response, error) {
	switch {
	case request.Method == "POST" || request.Method == "PUT": // create or update
		return transport.decorateConfigWriteOperation(request)
	default:
		return transport.executeKubernetesRequest(request, true)
	}
}

func (transport *baseTransport) decorateConfigWriteOperation(request *http.Request) (*http.Response, error) {
	body, err := utils.CopyBody(request)
	if err != nil {
		logrus.WithError(err).Debugf("[k8s configmap] failed parsing body")
	}

	response, err := transport.executeKubernetesRequest(request, false)

	if err == nil && (200 <= response.StatusCode && response.StatusCode < 300) {
		transport.logCreateConfigOperation(request, body)
	}

	return response, err
}

func (transport *baseTransport) logCreateConfigOperation(request *http.Request, body []byte) {
	cleanBody, err := hideConfigInfo(body)
	if err != nil {
		logrus.WithError(err).Debugf("[http,docker,config] message: failed cleaning request body")

	}
	if cleanBody == nil {
		cleanBody = make(map[string]interface{})
	}

	useractivityhttp.LogHttpActivity(transport.userActivityStore, transport.endpoint.Name, request, cleanBody)
}

// hideConfigInfo removes the confidential properties from the secret payload and returns the new payload
// it will read the request body and recreate it
func hideConfigInfo(body []byte) (interface{}, error) {
	type requestPayload struct {
		Metadata   interface{}       `json:"metadata"`
		Data       map[string]string `json:"data"`
		StringData map[string]string `json:"stringData"`
		BinaryData interface{}       `json:"binaryData"`
	}

	var payload requestPayload
	err := json.Unmarshal(body, &payload)
	if err != nil {
		return nil, errors.Wrap(err, "[configmaps] failed parsing body")
	}

	for key := range payload.Data {
		payload.Data[key] = useractivity.RedactedValue
	}

	for key := range payload.StringData {
		payload.StringData[key] = useractivity.RedactedValue
	}

	payload.BinaryData = useractivity.RedactedValue

	return payload, nil
}
