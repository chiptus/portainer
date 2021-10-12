package kubernetes

import (
	"net/http"

	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	useractivityhttp "github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/useractivity"
	"github.com/sirupsen/logrus"
)

func (transport *baseTransport) proxySecretsRequest(request *http.Request, namespace, requestPath string) (*http.Response, error) {
	switch request.Method {
	case "POST":
		return transport.proxySecretChangeOperation(request)
	case "GET":
		return transport.executeKubernetesRequest(request, false)
	case "PUT":
		return transport.proxySecretChangeOperation(request)
	case "DELETE":
		return transport.executeKubernetesRequest(request, true)
	default:
		return transport.executeKubernetesRequest(request, true)
	}
}

func (transport *baseTransport) proxySecretChangeOperation(request *http.Request) (*http.Response, error) {
	response, err := transport.executeKubernetesRequest(request, false)

	if err == nil && (200 <= response.StatusCode && response.StatusCode < 300) {
		body, _ := utils.GetRequestAsMap(request)
		transport.logWriteSecretsOperation(request, body)
	}

	return response, err
}

func (transport *baseTransport) logWriteSecretsOperation(request *http.Request, body map[string]interface{}) {
	cleanBody, err := hideSecretsInfo(body)
	if err != nil {
		logrus.WithError(err).Debugf("[http,docker,secret] message: failed cleaning request body")
	}

	useractivityhttp.LogHttpActivity(transport.userActivityStore, transport.endpoint.Name, request, cleanBody)
}

// hideSecretsInfo removes the confidential properties from the secret payload and returns the new payload
// it will read the request body and recreate it
func hideSecretsInfo(body map[string]interface{}) (interface{}, error) {
	if body == nil {
		return nil, nil
	}
	data := utils.GetJSONObject(body, "data")
	if data != nil {
		body["Data"] = useractivity.RedactedValue
	}

	stringData := utils.GetJSONObject(body, "stringData")
	if stringData != nil {
		for key := range stringData {
			stringData[key] = useractivity.RedactedValue
		}

		body["stringData"] = stringData
	}

	return body, nil
}
