package security

import (
	"net/http"
	"path"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

func getPortainerOperationAuthorization(url, method string) portainer.Authorization {
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
		return portaineree.OperationPortainerMOTD // all roles have access
	case "extensions":
		return portainerExtensionOperationAuthorization(url, method)
	case "registries":
		return portainerRegistryOperationAuthorization(url, method)
	case "resource_controls":
		return portainerResourceControlOperationAuthorization(url, method)
	case "roles":
		return portainerRoleOperationAuthorization(url, method)
	case "edge_update_schedules":
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

func portainerDockerhubOperationAuthorization(url, method string) portainer.Authorization {
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

func portainerEndpointGroupOperationAuthorization(url, method string) portainer.Authorization {
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

func portainerEndpointOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("endpoints", url)
	urlParts := strings.Split(url, "/")
	baseResource := strings.Split(urlParts[1], "?")[0]

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" { // GET /endpoints
			return portaineree.OperationPortainerEndpointList // all roles
		} else if resource != "" && action == "" { // GET /endpoints/:id
			return portaineree.OperationPortainerEndpointInspect // all roles
		}
		if action == "dockerhub" {
			return portainerDockerhubOperationAuthorization(strings.TrimPrefix(url, "/"+baseResource), method)
		}

		path, base := path.Split(action)
		if path == "" && base == "registries" { // GET /endpoints/:id/registries
			return portaineree.OperationPortainerRegistryList // all roles
		} else if path == "registries/" && base != "" { // GET /endpoints/:id/registries/:id
			return portaineree.OperationPortainerRegistryInspect // all roles
		}

	case http.MethodPost:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerEndpointCreate // unset - portainer admin only
			} else if resource == "snapshot" {
				return portaineree.OperationPortainerEndpointSnapshots // unset - portainer admin only
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

func portainerExtensionOperationAuthorization(url, method string) portainer.Authorization {
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

func portainerRegistryOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("registries", url)

	isInternalOperation := strings.HasPrefix(action, "v2/") ||
		strings.HasPrefix(action, "proxies/gitlab/") ||
		strings.HasPrefix(action, "ecr/")

	switch method {
	case http.MethodGet:
		// both are given to all roles but it doesn't make sense as it is applied on endpoint authorizations
		// this function is never called in sub endpoints routes (/api/endpoints/{ID}/registries)
		if resource == "" && action == "" {
			return portaineree.OperationPortainerRegistryList
		} else if (resource != "" && action == "") || isInternalOperation {
			return portaineree.OperationPortainerRegistryInspect
		}
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerRegistryCreate // unset - only portainer admin
		} else if action == "configure" {
			return portaineree.OperationPortainerRegistryConfigure // unset - only portainer admin
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerRegistryUpdate // unset - only portainer admin
		} else if isInternalOperation {
			return portaineree.OperationPortainerRegistryInternalUpdate // portainer admin + endpoint admin + standard user == update tag or repository
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerRegistryDelete // unset - only portainer admin
		} else if isInternalOperation {
			return portaineree.OperationPortainerRegistryInternalDelete // portainer admin + endpoint admin + standard user == delete tag or repository
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerResourceControlOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("resource_controls", url)

	switch method {
	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerResourceControlCreate // admin + edge admin + endpoint admin
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerResourceControlUpdate // admin + edge admin + endpoint admin + standard user
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerResourceControlDelete // admin + edge admin
		}
	}
	return portaineree.OperationPortainerUndefined
}

func portainerRoleOperationAuthorization(url, method string) portainer.Authorization {
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
func portainerScheduleOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("edge_update_schedules", url)

	// all are unset/unused = restricted to admins & edge admins

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

func portainerSettingsOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("settings", url)

	switch method {
	case http.MethodGet:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerSettingsInspect // unused/unset
		}
	case http.MethodPut:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerSettingsUpdate // unused/unset
			}
		case "checkLDAP": // this route pattern doesn't exist, maybe it's POST /api/ldap/check ?
			return portaineree.OperationPortainerSettingsLDAPCheck // unused/unset
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerStackOperationAuthorization(url, method string) portainer.Authorization {
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
		default:
			if resource == "create" {
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

func portainerTagOperationAuthorization(url, method string) portainer.Authorization {
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

func portainerTemplatesOperationAuthorization(url, method string) portainer.Authorization {
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
func portainerUploadOperationAuthorization(url, method string) portainer.Authorization {
	resource, _ := extractResourceAndActionFromURL("upload", url)

	switch method {
	case http.MethodPost:
		if resource == "tls" {
			return portaineree.OperationPortainerUploadTLS
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerUserOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("users", url)

	switch method {
	case http.MethodGet:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerUserList // all roles - DefaultPortainerAuthorizations()
			} else {
				return portaineree.OperationPortainerUserInspect // all roles - DefaultPortainerAuthorizations()
			}
		case "memberships":
			return portaineree.OperationPortainerUserMemberships // all roles - DefaultPortainerAuthorizations()
		case "tokens":
			return portaineree.OperationPortainerUserListToken // all roles - DefaultPortainerAuthorizations()
		case "gitcredentials":
			if resource == "" {
				return portaineree.OperationPortainerUserListGitCredential // all roles - DefaultPortainerAuthorizations()
			} else {
				return portaineree.OperationPortainerUserInspectGitCredential // all roles - DefaultPortainerAuthorizations()
			}
		}

	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerUserCreate // admin only
		} else if action == "tokens" {
			return portaineree.OperationPortainerUserCreateToken // all roles - DefaultPortainerAuthorizations()
		} else if action == "gitcredentials" {
			return portaineree.OperationPortainerUserCreateGitCredential // all roles - DefaultPortainerAuthorizations()
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerUserUpdate // admin only
		} else if resource != "" && action == "passwd" {
			return portaineree.OperationPortainerUserUpdatePassword // admin only
		} else if resource != "" && strings.HasPrefix(action, "gitcredentials") {
			return portaineree.OperationPortainerUserUpdateGitCredential // all roles - DefaultPortainerAuthorizations()
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerUserDelete // admin only
		} else if strings.HasPrefix(action, "tokens") {
			return portaineree.OperationPortainerUserRevokeToken // all roles - DefaultPortainerAuthorizations()
		} else if strings.HasPrefix(action, "gitcredentials") {
			return portaineree.OperationPortainerUserDeleteGitCredential // all roles - DefaultPortainerAuthorizations()
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerTeamOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("teams", url)

	switch method {
	case http.MethodGet:
		switch action {
		case "":
			if resource == "" {
				return portaineree.OperationPortainerTeamList // all roles - DefaultPortainerAuthorizations()
			} else {
				return portaineree.OperationPortainerTeamInspect // admin only
			}
		case "memberships":
			return portaineree.OperationPortainerTeamMemberships // admin only
		}

	case http.MethodPost:
		if resource == "" && action == "" {
			return portaineree.OperationPortainerTeamCreate // admin only
		}
	case http.MethodPut:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTeamUpdate // admin only
		}
	case http.MethodDelete:
		if resource != "" && action == "" {
			return portaineree.OperationPortainerTeamDelete // admin only
		}
	}

	return portaineree.OperationPortainerUndefined
}

func portainerTeamMembershipOperationAuthorization(url, method string) portainer.Authorization {
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

func portainerWebsocketOperationAuthorization(url, method string) portainer.Authorization {
	resource, _ := extractResourceAndActionFromURL("websocket", url)

	if resource == "exec" || resource == "attach" {
		return portaineree.OperationPortainerWebsocketExec
	}

	return portaineree.OperationPortainerUndefined
}

func portainerWebhookOperationAuthorization(url, method string) portainer.Authorization {
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
