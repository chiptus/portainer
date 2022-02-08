package security

import (
	"net/http"
	"path"
	"regexp"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
)

// IsAdminOrEndpointAdmin checks if current request is for an admin or an environment(endpoint) admin
func IsAdminOrEndpointAdmin(request *http.Request, dataStore portaineree.DataStore, endpointID portaineree.EndpointID) (bool, error) {
	tokenData, err := RetrieveTokenData(request)
	if err != nil {
		return false, err
	}

	if tokenData.Role == portaineree.AdministratorRole {
		return true, nil
	}

	user, err := dataStore.User().User(tokenData.ID)
	if err != nil {
		return false, err
	}

	_, endpointResourceAccess := user.EndpointAuthorizations[endpointID][portaineree.EndpointResourcesAccess]

	return endpointResourceAccess, nil
}

// AuthorizedOperation checks if operations is authorized
func authorizedOperation(operation *portaineree.APIOperationAuthorizationRequest) bool {
	operationAuthorization := getOperationAuthorization(operation.Path, operation.Method)
	return operation.Authorizations[operationAuthorization]
}

var dockerRule = regexp.MustCompile(`/(?P<identifier>\d+)/docker(?P<operation>/.*)`)
var k8sProxyRule = regexp.MustCompile(`/(?P<identifier>\d+)/kubernetes(?P<operation>/.*)`)
var k8sRule = regexp.MustCompile(`/kubernetes/(?P<identifier>\d+)(?P<operation>/.*)`)
var azureRule = regexp.MustCompile(`/(?P<identifier>\d+)/azure(?P<operation>/.*)`)

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

func getOperationAuthorization(url, method string) portaineree.Authorization {
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
	}

	return getPortainerOperationAuthorization(url, method)
}

func getKubernetesOperationAuthorization(url, method string) portaineree.Authorization {
	urlParts := strings.Split(url, "/")
	baseResource := strings.Split(urlParts[1], "?")[0]
	_, action := extractResourceAndActionFromURL(baseResource, url)

	authorizationsBindings := map[string]map[string]map[string]portaineree.Authorization{
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

func getPortainerOperationAuthorization(url, method string) portaineree.Authorization {
	urlParts := strings.Split(url, "/")
	baseResource := strings.Split(urlParts[1], "?")[0]

	switch baseResource {
	case "dockerhub":
		return portainerDockerhubOperationAuthorization(url, method)
	case "endpoint_groups":
		return portainerEndpointGroupOperationAuthorization(url, method)
	case "endpoints":
		return portainerEndpointOperationAuthorization(url, method)
	case "motd":
		return portaineree.OperationPortainerMOTD
	case "extensions":
		return portainerExtensionOperationAuthorization(url, method)
	case "registries":
		return portainerRegistryOperationAuthorization(url, method)
	case "resource_controls":
		return portainerResourceControlOperationAuthorization(url, method)
	case "roles":
		return portainerRoleOperationAuthorization(url, method)
	case "schedules":
		return portainerScheduleOperationAuthorization(url, method)
	case "settings":
		return portainerSettingsOperationAuthorization(url, method)
	case "stacks":
		return portainerStackOperationAuthorization(url, method)
	case "tags":
		return portainerTagOperationAuthorization(url, method)
	case "templates":
		return portainerTemplatesOperationAuthorization(url, method)
	case "upload":
		return portainerUploadOperationAuthorization(url, method)
	case "users":
		return portainerUserOperationAuthorization(url, method)
	case "teams":
		return portainerTeamOperationAuthorization(url, method)
	case "team_memberships":
		return portainerTeamMembershipOperationAuthorization(url, method)
	case "websocket":
		return portainerWebsocketOperationAuthorization(url, method)
	case "webhooks":
		return portainerWebhookOperationAuthorization(url, method)
	}

	return portaineree.OperationPortainerUndefined
}

func getDockerOperationAuthorization(url, method string) portaineree.Authorization {
	urlParts := strings.Split(url, "/")
	baseResource := strings.Split(urlParts[1], "?")[0]

	switch baseResource {
	case "v2":
		return getDockerOperationAuthorization(strings.TrimPrefix(url, "/"+baseResource), method)
	case "ping":
		return portaineree.OperationDockerAgentPing
	case "agents":
		return agentAgentsOperationAuthorization(url, method)
	case "browse":
		return agentBrowseOperationAuthorization(url, method)
	case "host":
		return agentHostOperationAuthorization(url, method)
	case "containers":
		return dockerContainerOperationAuthorization(url, method)
	case "images":
		return dockerImageOperationAuthorization(url, method)
	case "networks":
		return dockerNetworkOperationAuthorization(url, method)
	case "volumes":
		return dockerVolumeOperationAuthorization(url, method)
	case "exec":
		return dockerExecOperationAuthorization(url, method)
	case "swarm":
		return dockerSwarmOperationAuthorization(url, method)
	case "nodes":
		return dockerNodeOperationAuthorization(url, method)
	case "services":
		return dockerServiceOperationAuthorization(url, method)
	case "secrets":
		return dockerSecretOperationAuthorization(url, method)
	case "configs":
		return dockerConfigOperationAuthorization(url, method)
	case "tasks":
		return dockerTaskOperationAuthorization(url, method)
	case "plugins":
		return dockerPluginOperationAuthorization(url, method)
	case "info":
		return portaineree.OperationDockerInfo
	case "_ping":
		return portaineree.OperationDockerPing
	case "version":
		return portaineree.OperationDockerVersion
	case "events":
		return portaineree.OperationDockerEvents
	case "system/df": // TODO: this just cannot happen after strings.Split(url, "/"), can we use system instead?
		return portaineree.OperationDockerSystem
	case "session":
		return dockerSessionOperationAuthorization(url, method)
	case "distribution":
		return dockerDistributionOperationAuthorization(url, method)
	case "commit":
		return dockerCommitOperationAuthorization(url, method)
	case "build":
		return dockerBuildOperationAuthorization(url, method)
	default:
		return portaineree.OperationDockerUndefined
	}
}

func portainerDockerhubOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("dockerhub", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerDockerHubInspect
		}
	case http.MethodPut:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerDockerHubUpdate
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerEndpointGroupOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("endpoint_groups", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerEndpointGroupList
		} else if resource != "" && action == "" {
			return portaineree.OperationPortainerEndpointGroupInspect
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerEndpointGroupCreate
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerEndpointGroupUpdate
		} else if action == "access" {
			return portaineree.OperationPortainerEndpointGroupAccess
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerEndpointGroupDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerEndpointOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("endpoints", url)
	urlParts := strings.Split(url, "/")
	baseResource := strings.Split(urlParts[1], "?")[0]

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" { // GET /endpoints
			return portaineree.OperationPortainerEndpointList
		} else if resource != "" && action == "" { // GET /endpoints/:id
			return portaineree.OperationPortainerEndpointInspect
		}
		if action == "dockerhub" {
			return portainerDockerhubOperationAuthorization(strings.TrimPrefix(url, "/"+baseResource), method)
		}

		path, base := path.Split(action)
		if path == "" && base == "registries" { // GET /endpoints/:id/registries
			return portaineree.OperationPortainerRegistryList
		} else if path == "registries/" && base != "" { // GET /endpoints/:id/registries/:id
			return portaineree.OperationPortainerRegistryInspect
		}

	case http.MethodPost:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerEndpointCreate
			} else if resource == "snapshot" {
				return portaineree.OperationPortainerEndpointSnapshots
			}
		case "job":
			return portaineree.OperationPortainerEndpointJob
		case "snapshot":
			if resource != "" {
				return portaineree.OperationPortainerEndpointSnapshot
			}
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerEndpointUpdate
		} else if action == "access" {
			return portaineree.OperationPortainerEndpointUpdateAccess
		} else if action == "settings" {
			return portaineree.OperationPortainerEndpointUpdateSettings
		} else if action == "forceupdateservice" {
			return portaineree.OperationDockerServiceForceUpdateService
		} else if strings.HasPrefix(action, "registries/") {
			return portaineree.OperationPortainerRegistryUpdateAccess
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerEndpointDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerExtensionOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("extensions", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerExtensionList
		} else if resource != "" && action == "" {
			return portaineree.OperationPortainerExtensionInspect
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerExtensionCreate
		} else if action == "update" {
			return portaineree.OperationPortainerExtensionUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerExtensionDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerRegistryOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("registries", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerRegistryList
		} else if resource != "" && action == "" {
			return portaineree.OperationPortainerRegistryInspect
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerRegistryCreate
		} else if action == "configure" {
			return portaineree.OperationPortainerRegistryConfigure
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerRegistryUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerRegistryDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerResourceControlOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("resource_controls", url)

	switch method {
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerResourceControlCreate
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerResourceControlUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerResourceControlDelete
		}
	}
	return portaineree.OperationPortainerUndefined
}

func portainerRoleOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("roles", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerRoleList
		} else if resource != "" && action == "" {
			return portaineree.OperationPortainerRoleInspect
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerRoleCreate
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerRoleUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerRoleDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerScheduleOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("schedules", url)

	switch method {
	case http.MethodGet:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerScheduleList
			} else {
				return portaineree.OperationPortainerScheduleInspect
			}
		case "file":
			return portaineree.OperationPortainerScheduleFile
		case "tasks":
			return portaineree.OperationPortainerScheduleTasks
		}

	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerScheduleCreate
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerScheduleUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerScheduleDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerSettingsOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("settings", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerSettingsInspect
		}
	case http.MethodPut:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerSettingsUpdate
			}
		case "checkLDAP":
			return portaineree.OperationPortainerSettingsLDAPCheck
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerStackOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("stacks", url)

	switch method {
	case http.MethodGet:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerStackList
			} else {
				return portaineree.OperationPortainerStackInspect
			}
		case "file":
			return portaineree.OperationPortainerStackFile
		}

	case http.MethodPost:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerStackCreate
			}
		case "git", "stop", "start":
			return portaineree.OperationPortainerStackUpdate
		case "migrate":
			return portaineree.OperationPortainerStackMigrate
		}

	case http.MethodPut:
		if resource != "" && (action == "" || action == "git" || action == `git/redeploy`) {
			return portaineree.OperationPortainerStackUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerStackDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerTagOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("tags", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTagList
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTagCreate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTagDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerTeamMembershipOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("team_memberships", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTeamMembershipList
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTeamMembershipCreate
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTeamMembershipUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTeamMembershipDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerTeamOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("teams", url)

	switch method {
	case http.MethodGet:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerTeamList
			} else {
				return portaineree.OperationPortainerTeamInspect
			}
		case "memberships":
			return portaineree.OperationPortainerTeamMemberships
		}

	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTeamCreate
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTeamUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTeamDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerTemplatesOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("templates", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTemplateList
		} else if resource != "" && action == "" {
			return portaineree.OperationPortainerTemplateInspect
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTemplateCreate
		}
		if resource == "file" && action == "" {
			return portaineree.OperationPortainerTemplateInspect
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTemplateUpdate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTemplateDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerUploadOperationAuthorization(url, method string) portaineree.Authorization {
	resource, _ := extractResourceAndActionFromURL("upload", url)

	switch method {
	case http.MethodPost:
		if resource == "tls" {
			return portaineree.OperationPortainerUploadTLS
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerUserOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("users", url)

	switch method {
	case http.MethodGet:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerUserList
			} else {
				return portaineree.OperationPortainerUserInspect
			}
		case "memberships":
			return portaineree.OperationPortainerUserMemberships
		case "tokens":
			return portaineree.OperationPortainerUserListToken
		}

	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerUserCreate
		} else if action == "tokens" {
			return portaineree.OperationPortainerUserCreateToken
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerUserUpdate
		} else if resource != "" && action == "passwd" {
			return portaineree.OperationPortainerUserUpdatePassword
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerUserDelete
		} else if strings.HasPrefix(action, "tokens") {
			return portaineree.OperationPortainerUserRevokeToken
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerWebsocketOperationAuthorization(url, method string) portaineree.Authorization {
	resource, _ := extractResourceAndActionFromURL("websocket", url)

	if resource == "exec" || resource == "attach" {
		return portaineree.OperationPortainerWebsocketExec
	}

	return portaineree.OperationPortainerUndefined
}

func portainerWebhookOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("webhooks", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerWebhookList
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerWebhookCreate
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerWebhookDelete
		}
	}

	return portaineree.OperationPortainerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/network/network.go#L29
func dockerNetworkOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("networks", url)

	switch method {
	case http.MethodGet:
		// GET
		//router.NewGetRoute("/networks", r.getNetworksList),
		//router.NewGetRoute("/networks/", r.getNetworksList),
		//router.NewGetRoute("/networks/{id:.+}", r.getNetwork),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerNetworkList
			} else {
				return portaineree.OperationDockerNetworkInspect
			}
		}
	case http.MethodPost:
		//router.NewPostRoute("/networks/create", r.postNetworkCreate),
		//router.NewPostRoute("/networks/{id:.*}/connect", r.postNetworkConnect),
		//router.NewPostRoute("/networks/{id:.*}/disconnect", r.postNetworkDisconnect),
		//router.NewPostRoute("/networks/prune", r.postNetworksPrune),
		switch action {
		case "":
			if resource == "create" {
				return portaineree.OperationDockerNetworkCreate
			} else if resource == "prune" {
				return portaineree.OperationDockerNetworkPrune
			}
		case "connect":
			return portaineree.OperationDockerNetworkConnect
		case "disconnect":
			return portaineree.OperationDockerNetworkDisconnect
		}
	case http.MethodDelete:
		// DELETE
		// 	router.NewDeleteRoute("/networks/{id:.*}", r.deleteNetwork),
		if resource != "" && action == "" {
			return portaineree.OperationDockerNetworkDelete
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/volume/volume.go#L25
func dockerVolumeOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("volumes", url)

	switch method {
	case http.MethodGet:
		// GET
		//router.NewGetRoute("/volumes", r.getVolumesList),
		//	router.NewGetRoute("/volumes/{name:.*}", r.getVolumeByName),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerVolumeList
			} else {
				return portaineree.OperationDockerVolumeInspect
			}
		}
	case http.MethodPost:
		//router.NewPostRoute("/volumes/create", r.postVolumesCreate),
		//	router.NewPostRoute("/volumes/prune", r.postVolumesPrune),
		switch action {
		case "":
			if resource == "create" {
				return portaineree.OperationDockerVolumeCreate
			} else if resource == "prune" {
				return portaineree.OperationDockerVolumePrune
			}
		}
	case http.MethodDelete:
		// DELETE
		//router.NewDeleteRoute("/volumes/{name:.*}", r.deleteVolumes),
		if resource != "" && action == "" {
			return portaineree.OperationDockerVolumeDelete
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/container/container.go#L31
func dockerExecOperationAuthorization(url, method string) portaineree.Authorization {
	_, action := extractResourceAndActionFromURL("exec", url)

	switch method {
	case http.MethodGet:
		// GET
		// 		router.NewGetRoute("/exec/{id:.*}/json", r.getExecByID),
		if action == "json" {
			return portaineree.OperationDockerExecInspect
		}
	case http.MethodPost:
		// POST
		//router.NewPostRoute("/exec/{name:.*}/start", r.postContainerExecStart),
		//	router.NewPostRoute("/exec/{name:.*}/resize", r.postContainerExecResize),
		if action == "start" {
			return portaineree.OperationDockerExecStart
		} else if action == "resize" {
			return portaineree.OperationDockerExecResize
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/swarm/cluster.go#L25
func dockerSwarmOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("swarm", url)

	switch method {
	case http.MethodGet:
		// GET
		//	router.NewGetRoute("/swarm", sr.inspectCluster),
		//	router.NewGetRoute("/swarm/unlockkey", sr.getUnlockKey),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerSwarmInspect
			} else {
				return portaineree.OperationDockerSwarmUnlockKey
			}
		}
	case http.MethodPost:
		// POST
		//router.NewPostRoute("/swarm/init", sr.initCluster),
		//	router.NewPostRoute("/swarm/join", sr.joinCluster),
		//	router.NewPostRoute("/swarm/leave", sr.leaveCluster),
		//	router.NewPostRoute("/swarm/update", sr.updateCluster),
		//	router.NewPostRoute("/swarm/unlock", sr.unlockCluster),
		switch action {
		case "":
			switch resource {
			case "init":
				return portaineree.OperationDockerSwarmInit
			case "join":
				return portaineree.OperationDockerSwarmJoin
			case "leave":
				return portaineree.OperationDockerSwarmLeave
			case "update":
				return portaineree.OperationDockerSwarmUpdate
			case "unlock":
				return portaineree.OperationDockerSwarmUnlock
			}
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/swarm/cluster.go#L25
func dockerNodeOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("nodes", url)

	switch method {
	case http.MethodGet:
		// GET
		//router.NewGetRoute("/nodes", sr.getNodes),
		//	router.NewGetRoute("/nodes/{id}", sr.getNode),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerNodeList
			} else {
				return portaineree.OperationDockerNodeInspect
			}
		}
	case http.MethodPost:
		// POST
		//	router.NewPostRoute("/nodes/{id}/update", sr.updateNode)
		if action == "update" {
			return portaineree.OperationDockerNodeUpdate
		}
	case http.MethodDelete:
		// DELETE
		//	router.NewDeleteRoute("/nodes/{id}", sr.removeNode),
		if resource != "" {
			return portaineree.OperationDockerNodeDelete
		}

	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/swarm/cluster.go#L25
func dockerServiceOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("services", url)

	switch method {
	case http.MethodGet:
		//// GET
		//router.NewGetRoute("/services", sr.getServices),
		//	router.NewGetRoute("/services/{id}", sr.getService),
		//	router.NewGetRoute("/services/{id}/logs", sr.getServiceLogs),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerServiceList
			} else {
				return portaineree.OperationDockerServiceInspect
			}
		case "logs":
			return portaineree.OperationDockerServiceLogs
		}
	case http.MethodPost:
		//// POST
		//	router.NewPostRoute("/services/create", sr.createService),
		//	router.NewPostRoute("/services/{id}/update", sr.updateService),
		switch action {
		case "":
			if resource == "create" {
				return portaineree.OperationDockerServiceCreate
			}
		case "update":
			return portaineree.OperationDockerServiceUpdate
		}
	case http.MethodDelete:
		//// DELETE
		//	router.NewDeleteRoute("/services/{id}", sr.removeService),
		if resource != "" && action == "" {
			return portaineree.OperationDockerServiceDelete
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/swarm/cluster.go#L25
func dockerSecretOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("secrets", url)

	switch method {
	case http.MethodGet:
		//// GET
		//router.NewGetRoute("/secrets", sr.getSecrets),
		//	router.NewGetRoute("/secrets/{id}", sr.getSecret),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerSecretList
			} else {
				return portaineree.OperationDockerSecretInspect
			}
		}
	case http.MethodPost:
		//// POST
		//	router.NewPostRoute("/secrets/create", sr.createSecret),
		//	router.NewPostRoute("/secrets/{id}/update", sr.updateSecret),
		switch action {
		case "":
			if resource == "create" {
				return portaineree.OperationDockerSecretCreate
			}
		case "update":
			return portaineree.OperationDockerSecretUpdate
		}
	case http.MethodDelete:
		//// DELETE
		//	router.NewDeleteRoute("/secrets/{id}", sr.removeSecret),
		if resource != "" && action == "" {
			return portaineree.OperationDockerSecretDelete
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/swarm/cluster.go#L25
func dockerConfigOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("configs", url)

	switch method {
	case http.MethodGet:
		//// GET
		//router.NewGetRoute("/configs", sr.getConfigs),
		//	router.NewGetRoute("/configs/{id}", sr.getConfig),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerConfigList
			} else {
				return portaineree.OperationDockerConfigInspect
			}
		}
	case http.MethodPost:
		//// POST
		//	router.NewPostRoute("/configs/create", sr.createConfig),
		//	router.NewPostRoute("/configs/{id}/update", sr.updateConfig),
		switch action {
		case "":
			if resource == "create" {
				return portaineree.OperationDockerConfigCreate
			}
		case "update":
			return portaineree.OperationDockerConfigUpdate
		}
	case http.MethodDelete:
		//// DELETE
		//	router.NewDeleteRoute("/configs/{id}", sr.removeConfig),
		if resource != "" && action == "" {
			return portaineree.OperationDockerConfigDelete
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/swarm/cluster.go#L25
func dockerTaskOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("tasks", url)

	switch method {
	case http.MethodGet:
		//// GET
		//router.NewGetRoute("/tasks", sr.getTasks),
		//	router.NewGetRoute("/tasks/{id}", sr.getTask),
		//	router.NewGetRoute("/tasks/{id}/logs", sr.getTaskLogs),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerTaskList
			} else {
				return portaineree.OperationDockerTaskInspect
			}
		case "logs":
			return portaineree.OperationDockerTaskLogs
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
//https://github.com/moby/moby/blob/c12f09bf99/api/server/router/plugin/plugin.go#L25
func dockerPluginOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("plugins", url)

	switch method {
	case http.MethodGet:
		//// GET
		//router.NewGetRoute("/plugins", r.listPlugins),
		//	router.NewGetRoute("/plugins/{name:.*}/json", r.inspectPlugin),
		//	router.NewGetRoute("/plugins/privileges", r.getPrivileges),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerPluginList
			} else if resource == "privileges" {
				return portaineree.OperationDockerPluginPrivileges
			}
		case "json":
			return portaineree.OperationDockerPluginInspect
		}
	case http.MethodPost:
		//// POST
		//	router.NewPostRoute("/plugins/pull", r.pullPlugin),
		//	router.NewPostRoute("/plugins/create", r.createPlugin),
		//	router.NewPostRoute("/plugins/{name:.*}/enable", r.enablePlugin),
		//	router.NewPostRoute("/plugins/{name:.*}/disable", r.disablePlugin),
		//	router.NewPostRoute("/plugins/{name:.*}/push", r.pushPlugin),
		//	router.NewPostRoute("/plugins/{name:.*}/upgrade", r.upgradePlugin),
		//	router.NewPostRoute("/plugins/{name:.*}/set", r.setPlugin),
		switch action {
		case "":
			if resource == "pull" {
				return portaineree.OperationDockerPluginPull
			} else if resource == "create" {
				return portaineree.OperationDockerPluginCreate
			}
		case "enable":
			return portaineree.OperationDockerPluginEnable
		case "disable":
			return portaineree.OperationDockerPluginDisable
		case "push":
			return portaineree.OperationDockerPluginPush
		case "upgrade":
			return portaineree.OperationDockerPluginUpgrade
		case "set":
			return portaineree.OperationDockerPluginSet
		}
	case http.MethodDelete:
		//// DELETE
		//	router.NewDeleteRoute("/plugins/{name:.*}", r.removePlugin),
		if resource != "" && action == "" {
			return portaineree.OperationDockerPluginDelete
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/session/session.go
func dockerSessionOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("session", url)

	switch method {
	case http.MethodPost:
		//// POST
		//router.NewPostRoute("/session", r.startSession),
		if action == "" && resource == "" {
			return portaineree.OperationDockerSessionStart
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/distribution/distribution.go#L26
func dockerDistributionOperationAuthorization(url, method string) portaineree.Authorization {
	_, action := extractResourceAndActionFromURL("distribution", url)

	switch method {
	case http.MethodGet:
		//// GET
		//router.NewGetRoute("/distribution/{name:.*}/json", r.getDistributionInfo),
		if action == "json" {
			return portaineree.OperationDockerDistributionInspect
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/container/container.go#L31
func dockerCommitOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("commit", url)

	switch method {
	case http.MethodPost:
		//// POST
		// router.NewPostRoute("/commit", r.postCommit),
		if resource == "" && action == "" {
			return portaineree.OperationDockerImageCommit
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/build/build.go#L32
func dockerBuildOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("build", url)

	switch method {
	case http.MethodPost:
		//// POST
		//	router.NewPostRoute("/build", r.postBuild),
		//	router.NewPostRoute("/build/prune", r.postPrune),
		//	router.NewPostRoute("/build/cancel", r.postCancel),
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationDockerImageBuild
			} else if resource == "prune" {
				return portaineree.OperationDockerBuildPrune
			} else if resource == "cancel" {
				return portaineree.OperationDockerBuildCancel
			}
		}
	}

	return portaineree.OperationDockerUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/image/image.go#L26
func dockerImageOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("images", url)

	switch method {
	case http.MethodGet:
		//// GET
		//router.NewGetRoute("/images/json", r.getImagesJSON),
		//	router.NewGetRoute("/images/search", r.getImagesSearch),
		//	router.NewGetRoute("/images/get", r.getImagesGet),
		//	router.NewGetRoute("/images/{name:.*}/get", r.getImagesGet),
		//	router.NewGetRoute("/images/{name:.*}/history", r.getImagesHistory),
		//	router.NewGetRoute("/images/{name:.*}/json", r.getImagesByName),
		switch action {
		case "":
			if resource == "json" {
				return portaineree.OperationDockerImageList
			} else if resource == "search" {
				return portaineree.OperationDockerImageSearch
			} else if resource == "get" {
				return portaineree.OperationDockerImageGetAll
			}
		case "get":
			return portaineree.OperationDockerImageGet
		case "history":
			return portaineree.OperationDockerImageHistory
		case "json":
			return portaineree.OperationDockerImageInspect
		}
	case http.MethodPost:
		//// POST
		//	router.NewPostRoute("/images/load", r.postImagesLoad),
		//	router.NewPostRoute("/images/create", r.postImagesCreate),
		//	router.NewPostRoute("/images/{name:.*}/push", r.postImagesPush),
		//	router.NewPostRoute("/images/{name:.*}/tag", r.postImagesTag),
		//	router.NewPostRoute("/images/prune", r.postImagesPrune),
		switch action {
		case "":
			if resource == "load" {
				return portaineree.OperationDockerImageLoad
			} else if resource == "create" {
				return portaineree.OperationDockerImageCreate
			} else if resource == "prune" {
				return portaineree.OperationDockerImagePrune
			}
		case "push":
			return portaineree.OperationDockerImagePush
		case "tag":
			return portaineree.OperationDockerImageTag
		}
	case http.MethodDelete:
		//// DELETE
		//	router.NewDeleteRoute("/images/{name:.*}", r.deleteImages)
		return portaineree.OperationDockerImageDelete
	}

	return portaineree.OperationDockerUndefined
}

func agentAgentsOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("agents", url)

	switch method {
	case http.MethodGet:
		if action == "" && resource == "" {
			return portaineree.OperationDockerAgentList
		}
	}

	return portaineree.OperationDockerAgentUndefined
}

func agentBrowseOperationAuthorization(url, method string) portaineree.Authorization {
	resource, _ := extractResourceAndActionFromURL("browse", url)

	switch method {
	case http.MethodGet:
		switch resource {
		case "ls":
			return portaineree.OperationDockerAgentBrowseList
		case "get":
			return portaineree.OperationDockerAgentBrowseGet
		}
	case http.MethodDelete:
		if resource == "delete" {
			return portaineree.OperationDockerAgentBrowseDelete
		}
	case http.MethodPut:
		if resource == "rename" {
			return portaineree.OperationDockerAgentBrowseRename
		}
	case http.MethodPost:
		if resource == "put" {
			return portaineree.OperationDockerAgentBrowsePut
		}

	}

	return portaineree.OperationDockerAgentUndefined
}

func agentHostOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("host", url)

	switch method {
	case http.MethodGet:
		if action == "" && resource == "info" {
			return portaineree.OperationDockerAgentHostInfo
		}
	}

	return portaineree.OperationDockerAgentUndefined
}

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/container/container.go#L31
func dockerContainerOperationAuthorization(url, method string) portaineree.Authorization {
	resource, action := extractResourceAndActionFromURL("containers", url)

	switch method {
	case http.MethodHead:
		//// HEAD
		//router.NewHeadRoute("/containers/{name:.*}/archive", r.headContainersArchive),
		if action == "archive" {
			return portaineree.OperationDockerContainerArchiveInfo
		}
	case http.MethodGet:
		//// GET
		//	router.NewGetRoute("/containers/json", r.getContainersJSON),
		//	router.NewGetRoute("/containers/{name:.*}/export", r.getContainersExport),
		//	router.NewGetRoute("/containers/{name:.*}/changes", r.getContainersChanges),
		//	router.NewGetRoute("/containers/{name:.*}/json", r.getContainersByName),
		//	router.NewGetRoute("/containers/{name:.*}/top", r.getContainersTop),
		//	router.NewGetRoute("/containers/{name:.*}/logs", r.getContainersLogs),
		//	router.NewGetRoute("/containers/{name:.*}/stats", r.getContainersStats),
		//	router.NewGetRoute("/containers/{name:.*}/attach/ws", r.wsContainersAttach),
		//	router.NewGetRoute("/exec/{id:.*}/json", r.getExecByID),
		//	router.NewGetRoute("/containers/{name:.*}/archive", r.getContainersArchive),
		switch action {
		case "":
			if resource == "json" {
				return portaineree.OperationDockerContainerList
			}
		case "export":
			return portaineree.OperationDockerContainerExport
		case "changes":
			return portaineree.OperationDockerContainerChanges
		case "json":
			return portaineree.OperationDockerContainerInspect
		case "top":
			return portaineree.OperationDockerContainerTop
		case "logs":
			return portaineree.OperationDockerContainerLogs
		case "stats":
			return portaineree.OperationDockerContainerStats
		case "attach/ws":
			return portaineree.OperationDockerContainerAttachWebsocket
		case "archive":
			return portaineree.OperationDockerContainerArchive
		}
	case http.MethodPost:
		//// POST
		//	router.NewPostRoute("/containers/create", r.postContainersCreate),
		//	router.NewPostRoute("/containers/{name:.*}/kill", r.postContainersKill),
		//	router.NewPostRoute("/containers/{name:.*}/pause", r.postContainersPause),
		//	router.NewPostRoute("/containers/{name:.*}/unpause", r.postContainersUnpause),
		//	router.NewPostRoute("/containers/{name:.*}/restart", r.postContainersRestart),
		//	router.NewPostRoute("/containers/{name:.*}/start", r.postContainersStart),
		//	router.NewPostRoute("/containers/{name:.*}/stop", r.postContainersStop),
		//	router.NewPostRoute("/containers/{name:.*}/wait", r.postContainersWait),
		//	router.NewPostRoute("/containers/{name:.*}/resize", r.postContainersResize),
		//	router.NewPostRoute("/containers/{name:.*}/attach", r.postContainersAttach),
		//	router.NewPostRoute("/containers/{name:.*}/copy", r.postContainersCopy), // Deprecated since 1.8, Errors out since 1.12
		//	router.NewPostRoute("/containers/{name:.*}/exec", r.postContainerExecCreate),
		//	router.NewPostRoute("/exec/{name:.*}/start", r.postContainerExecStart),
		//	router.NewPostRoute("/exec/{name:.*}/resize", r.postContainerExecResize),
		//	router.NewPostRoute("/containers/{name:.*}/rename", r.postContainerRename),
		//	router.NewPostRoute("/containers/{name:.*}/update", r.postContainerUpdate),
		//	router.NewPostRoute("/containers/prune", r.postContainersPrune),
		//	router.NewPostRoute("/commit", r.postCommit),
		switch action {
		case "":
			if resource == "create" {
				return portaineree.OperationDockerContainerCreate
			} else if resource == "prune" {
				return portaineree.OperationDockerContainerPrune
			}
		case "kill":
			return portaineree.OperationDockerContainerKill
		case "pause":
			return portaineree.OperationDockerContainerPause
		case "unpause":
			return portaineree.OperationDockerContainerUnpause
		case "restart":
			return portaineree.OperationDockerContainerRestart
		case "start":
			return portaineree.OperationDockerContainerStart
		case "stop":
			return portaineree.OperationDockerContainerStop
		case "wait":
			return portaineree.OperationDockerContainerWait
		case "resize":
			return portaineree.OperationDockerContainerResize
		case "attach":
			return portaineree.OperationDockerContainerAttach
		case "exec":
			return portaineree.OperationDockerContainerExec
		case "rename":
			return portaineree.OperationDockerContainerRename
		case "update":
			return portaineree.OperationDockerContainerUpdate
		}
	case http.MethodPut:
		//// PUT
		//	router.NewPutRoute("/containers/{name:.*}/archive", r.putContainersArchive),
		if action == "archive" {
			return portaineree.OperationDockerContainerPutContainerArchive
		}
	case http.MethodDelete:
		//// DELETE
		//	router.NewDeleteRoute("/containers/{name:.*}", r.deleteContainers),
		if resource != "" && action == "" {
			return portaineree.OperationDockerContainerDelete
		}
	}

	return portaineree.OperationDockerUndefined
}
