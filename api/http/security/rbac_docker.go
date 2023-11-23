package security

import (
	"net/http"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

func getDockerOperationAuthorization(url, method string) portainer.Authorization {
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

func agentAgentsOperationAuthorization(url, method string) portainer.Authorization {
	resource, action := extractResourceAndActionFromURL("agents", url)

	switch method {
	case http.MethodGet:
		if action == "" && resource == "" {
			return portaineree.OperationDockerAgentList
		}
	}

	return portaineree.OperationDockerAgentUndefined
}

func agentBrowseOperationAuthorization(url, method string) portainer.Authorization {
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

func agentHostOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerContainerOperationAuthorization(url, method string) portainer.Authorization {
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

// Based on the routes available at
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/network/network.go#L29
func dockerNetworkOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerVolumeOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerExecOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerSwarmOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerNodeOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerServiceOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerSecretOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerConfigOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerTaskOperationAuthorization(url, method string) portainer.Authorization {
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
// https://github.com/moby/moby/blob/c12f09bf99/api/server/router/plugin/plugin.go#L25
func dockerPluginOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerSessionOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerDistributionOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerCommitOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerBuildOperationAuthorization(url, method string) portainer.Authorization {
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
func dockerImageOperationAuthorization(url, method string) portainer.Authorization {
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
