package kubernetes

import (
	"log"
	"net/http"
	"path"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	"github.com/portainer/portainer/api/http/security"
	useractivityhttp "github.com/portainer/portainer/api/http/useractivity"
	"github.com/portainer/portainer/api/kubernetes/privateregistries"
	"github.com/portainer/portainer/api/useractivity"
	v1 "k8s.io/api/core/v1"
)

func (transport *baseTransport) proxySecretsRequest(request *http.Request, namespace, requestPath string) (*http.Response, error) {
	switch request.Method {
	case "POST":
		return transport.proxySecretCreationOperation(request)
	case "GET":
		if path.Base(requestPath) == "secrets" {
			return transport.proxySecretListOperation(request)
		}
		return transport.proxySecretInspectOperation(request)
	case "PUT":
		return transport.proxySecretUpdateOperation(request)
	case "DELETE":
		return transport.proxySecretDeleteOperation(request, namespace)
	default:
		return transport.executeKubernetesRequest(request, true)
	}
}

func (transport *baseTransport) proxySecretCreationOperation(request *http.Request) (*http.Response, error) {
	body, err := utils.GetRequestAsMap(request)
	if err != nil {
		return nil, err
	}

	if isSecretRepresentPrivateRegistry(body) {
		return utils.WriteAccessDeniedResponse()
	}

	err = utils.RewriteRequest(request, body)
	if err != nil {
		return nil, err
	}

	response, err := transport.executeKubernetesRequest(request, false)

	if err == nil && (200 <= response.StatusCode && response.StatusCode < 300) {
		transport.logWriteSecretsOperation(request, body)
	}

	return response, err
}

func (transport *baseTransport) logWriteSecretsOperation(request *http.Request, body map[string]interface{}) {
	cleanBody, err := hideSecretsInfo(body)
	if err != nil {
		log.Printf("[ERROR] [http,docker,config] [message: failed cleaning request body] [error: %s]", err)
		return
	}

	useractivityhttp.LogHttpActivity(transport.userActivityStore, transport.endpoint.Name, request, cleanBody)
}

// hideSecretsInfo removes the confidential properties from the secret payload and returns the new payload
// it will read the request body and recreate it
func hideSecretsInfo(body map[string]interface{}) (interface{}, error) {
	type requestPayload struct {
		Metadata   interface{}       `json:"metadata"`
		Data       map[string]string `json:"data"`
		StringData map[string]string `json:"stringData"`
		Type       string            `json:"type"`
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

func (transport *baseTransport) proxySecretListOperation(request *http.Request) (*http.Response, error) {
	response, err := transport.executeKubernetesRequest(request, false)
	if err != nil {
		return nil, err
	}

	canAccess, err := hasAuthorization(request, transport.dataStore.User(), transport.endpoint.ID, portainer.OperationK8sRegistrySecretList)
	if err != nil {
		return nil, err
	}

	if canAccess {
		return response, nil
	}

	body, err := utils.GetResponseAsJSONObject(response)
	if err != nil {
		return nil, err
	}

	items := utils.GetArrayObject(body, "items")

	if items == nil {
		utils.RewriteResponse(response, body, response.StatusCode)
		return response, nil
	}

	filteredItems := []interface{}{}
	for _, item := range items {
		itemObj := item.(map[string]interface{})
		if !isSecretRepresentPrivateRegistry(itemObj) {
			filteredItems = append(filteredItems, item)
		}
	}

	body["items"] = filteredItems

	utils.RewriteResponse(response, body, response.StatusCode)
	return response, nil
}

func (transport *baseTransport) proxySecretInspectOperation(request *http.Request) (*http.Response, error) {
	response, err := transport.executeKubernetesRequest(request, false)
	if err != nil {
		return nil, err
	}

	canAccess, err := hasAuthorization(request, transport.dataStore.User(), transport.endpoint.ID, portainer.OperationK8sRegistrySecretInspect)
	if err != nil {
		return nil, err
	}

	if canAccess {
		return response, nil
	}

	body, err := utils.GetResponseAsJSONObject(response)
	if err != nil {
		return nil, err
	}

	if isSecretRepresentPrivateRegistry(body) {
		return utils.WriteAccessDeniedResponse()
	}

	err = utils.RewriteResponse(response, body, response.StatusCode)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (transport *baseTransport) proxySecretUpdateOperation(request *http.Request) (*http.Response, error) {
	body, err := utils.GetRequestAsMap(request)
	if err != nil {
		return nil, err
	}

	if isSecretRepresentPrivateRegistry(body) {
		return utils.WriteAccessDeniedResponse()
	}

	err = utils.RewriteRequest(request, body)
	if err != nil {
		return nil, err
	}

	response, err := transport.executeKubernetesRequest(request, false)

	if err == nil && (200 <= response.StatusCode && response.StatusCode < 300) {
		transport.logWriteSecretsOperation(request, body)
	}

	return response, err
}

func (transport *baseTransport) proxySecretDeleteOperation(request *http.Request, namespace string) (*http.Response, error) {
	kcl, err := transport.k8sClientFactory.GetKubeClient(transport.endpoint)
	if err != nil {
		return nil, err
	}

	secretName := path.Base(request.RequestURI)

	isRegistrySecret, err := kcl.IsRegistrySecret(namespace, secretName)
	if err != nil {
		return nil, err
	}

	if isRegistrySecret {
		return utils.WriteAccessDeniedResponse()
	}

	return transport.executeKubernetesRequest(request, true)
}

func isSecretRepresentPrivateRegistry(secret map[string]interface{}) bool {
	if secret["type"].(string) != string(v1.SecretTypeDockerConfigJson) {
		return false
	}

	metadata := utils.GetJSONObject(secret, "metadata")
	annotations := utils.GetJSONObject(metadata, "annotations")
	_, ok := annotations[privateregistries.RegistryIDLabel]

	return ok
}

// hasAuthorization checks if current request is an admin or has the correct authorization
func hasAuthorization(request *http.Request, userService portainer.UserService, endpointID portainer.EndpointID, authorization portainer.Authorization) (bool, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return false, err
	}

	if tokenData.Role == portainer.AdministratorRole {
		return true, nil
	}

	user, err := userService.User(tokenData.ID)
	if err != nil {
		return false, err
	}

	return user.EndpointAuthorizations[endpointID][authorization], nil
}
