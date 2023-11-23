package security

import (
	"net/http"
	"regexp"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// AuthorizedOperation checks if operations is authorized
func authorizedOperation(operation *portaineree.APIOperationAuthorizationRequest) bool {
	operationAuthorization := getOperationAuthorization(operation.Path, operation.Method)
	return operation.Authorizations[operationAuthorization]
}

var dockerRule = regexp.MustCompile(`/(?P<identifier>\d+)/docker(?P<operation>/.*)`)
var k8sProxyRule = regexp.MustCompile(`/(?P<identifier>\d+)/kubernetes(?P<operation>/.*)`)
var k8sRule = regexp.MustCompile(`/kubernetes/(?P<identifier>\d+)(?P<operation>/.*)`)
var azureRule = regexp.MustCompile(`/(?P<identifier>\d+)/azure(?P<operation>/.*)`)
var agentRule = regexp.MustCompile(`/(?P<identifier>\d+)/agent(?P<operation>/.*)`)

// var cloudRule = regexp.MustCompile(`/cloud/endpoints/(?P<identifier>\d+)`)

func extractMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}

func extractResourceAndActionFromURL(routeResource, url string) (string, string) {
	routePattern := regexp.MustCompile(`/` + routeResource + `/(?P<resource>[^/?]*)/?(?P<action>[^?]*)?(\?.*)?`)
	urlComponents := extractMatches(routePattern, url)

	// TODO: optional log statement for debug
	//fmt.Printf("[DEBUG] - RBAC | OPERATION: %s | resource: %s | action: %s\n", url, urlComponents["resource"], urlComponents["action"])

	return urlComponents["resource"], urlComponents["action"]
}

func getOperationAuthorization(url, method string) portainer.Authorization {
	if dockerRule.MatchString(url) {
		match := dockerRule.FindStringSubmatch(url)
		return getDockerOperationAuthorization(strings.TrimPrefix(url, "/"+match[1]+"/docker"), method)
	} else if k8sProxyRule.MatchString(url) {
		// if the k8sProxyRule is matched, only tests if the user can access
		// the current environment(endpoint). The namespace + resource authorization
		// is done in the k8s level.
		return portaineree.OperationK8sResourcePoolsR
	} else if azureRule.MatchString(url) {
		match := azureRule.FindStringSubmatch(url)
		return getAzureOperationAuthorization(strings.TrimPrefix(url, "/"+match[1]+"/azure"), method)
	} else if k8sRule.MatchString(url) {
		match := k8sRule.FindStringSubmatch(url)
		return getKubernetesOperationAuthorization(strings.TrimPrefix(url, "/kubernetes/"+match[1]), method)
	} else if agentRule.MatchString(url) {
		match := agentRule.FindStringSubmatch(url)
		return portainerAgentOperationAuthorization(strings.TrimPrefix(url, "/"+match[1]+"/agent"), method)
	}

	return getPortainerOperationAuthorization(url, method)
}

func IsAdmin(role portainer.UserRole) bool {
	return role == portaineree.AdministratorRole
}

func IsAdminOrEdgeAdmin(role portainer.UserRole) bool {
	return role == portaineree.AdministratorRole || role == portaineree.EdgeAdminRole
}

func IsAdminContext(context *RestrictedRequestContext) bool {
	return context.IsAdmin
}

func IsAdminOrEdgeAdminContext(context *RestrictedRequestContext) bool {
	return context.IsAdmin || context.IsEdgeAdmin
}

// IsAdminOrEndpointAdmin checks if current request is for an admin, edge admin, or an environment(endpoint) admin
//
// EE-6176 TODO later: move this check to RBAC layer performed before in-handler execution (see usage references of this func)
func IsAdminOrEndpointAdmin(request *http.Request, tx dataservices.DataStoreTx, endpointID portainer.EndpointID) (bool, error) {
	tokenData, err := RetrieveTokenData(request)
	if err != nil {
		return false, err
	}

	if IsAdminOrEdgeAdmin(tokenData.Role) {
		return true, nil
	}

	user, err := tx.User().Read(tokenData.ID)
	if err != nil {
		return false, err
	}

	_, endpointResourceAccess := user.EndpointAuthorizations[endpointID][portaineree.EndpointResourcesAccess]

	return endpointResourceAccess, nil
}
