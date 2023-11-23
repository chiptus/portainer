package security

import (
	"net/http"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

func getKubernetesOperationAuthorization(url, method string) portainer.Authorization {
	urlParts := strings.Split(url, "/")
	baseResource := strings.Split(urlParts[1], "?")[0]
	_, action := extractResourceAndActionFromURL(baseResource, url)

	authorizationsBindings := map[string]map[string]map[string]portainer.Authorization{
		"namespaces": {
			"system": {
				http.MethodPut: portaineree.OperationK8sResourcePoolDetailsW,
			},
		},
	}

	if authorization, ok := authorizationsBindings[baseResource][action][method]; ok {
		return authorization
	}
	return portaineree.OperationK8sUndefined
}
