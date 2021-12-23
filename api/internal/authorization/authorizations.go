package authorization

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
)

// Service represents a service used to
// update authorizations associated to a user or team.
type (
	Service struct {
		dataStore         portaineree.DataStore
		authEventHandlers map[string]portaineree.AuthEventHandler
		K8sClientFactory  *cli.ClientFactory
	}
)

// NewService returns a point to a new Service instance.
func NewService(dataStore portaineree.DataStore) *Service {
	return &Service{
		dataStore:         dataStore,
		authEventHandlers: make(map[string]portaineree.AuthEventHandler),
	}
}

// DefaultEndpointAuthorizationsForEndpointAdministratorRole returns the default environment(endpoint) authorizations
// associated to the environment(endpoint) administrator role.
func DefaultEndpointAuthorizationsForEndpointAdministratorRole() portaineree.Authorizations {
	return unionAuthorizations(map[portaineree.Authorization]bool{
		portaineree.OperationDockerContainerArchiveInfo:         true,
		portaineree.OperationDockerContainerList:                true,
		portaineree.OperationDockerContainerExport:              true,
		portaineree.OperationDockerContainerChanges:             true,
		portaineree.OperationDockerContainerInspect:             true,
		portaineree.OperationDockerContainerTop:                 true,
		portaineree.OperationDockerContainerLogs:                true,
		portaineree.OperationDockerContainerStats:               true,
		portaineree.OperationDockerContainerAttachWebsocket:     true,
		portaineree.OperationDockerContainerArchive:             true,
		portaineree.OperationDockerContainerCreate:              true,
		portaineree.OperationDockerContainerPrune:               true,
		portaineree.OperationDockerContainerKill:                true,
		portaineree.OperationDockerContainerPause:               true,
		portaineree.OperationDockerContainerUnpause:             true,
		portaineree.OperationDockerContainerRestart:             true,
		portaineree.OperationDockerContainerStart:               true,
		portaineree.OperationDockerContainerStop:                true,
		portaineree.OperationDockerContainerWait:                true,
		portaineree.OperationDockerContainerResize:              true,
		portaineree.OperationDockerContainerAttach:              true,
		portaineree.OperationDockerContainerExec:                true,
		portaineree.OperationDockerContainerRename:              true,
		portaineree.OperationDockerContainerUpdate:              true,
		portaineree.OperationDockerContainerPutContainerArchive: true,
		portaineree.OperationDockerContainerDelete:              true,
		portaineree.OperationDockerImageList:                    true,
		portaineree.OperationDockerImageSearch:                  true,
		portaineree.OperationDockerImageGetAll:                  true,
		portaineree.OperationDockerImageGet:                     true,
		portaineree.OperationDockerImageHistory:                 true,
		portaineree.OperationDockerImageInspect:                 true,
		portaineree.OperationDockerImageLoad:                    true,
		portaineree.OperationDockerImageCreate:                  true,
		portaineree.OperationDockerImagePrune:                   true,
		portaineree.OperationDockerImagePush:                    true,
		portaineree.OperationDockerImageTag:                     true,
		portaineree.OperationDockerImageDelete:                  true,
		portaineree.OperationDockerImageCommit:                  true,
		portaineree.OperationDockerImageBuild:                   true,
		portaineree.OperationDockerNetworkList:                  true,
		portaineree.OperationDockerNetworkInspect:               true,
		portaineree.OperationDockerNetworkCreate:                true,
		portaineree.OperationDockerNetworkConnect:               true,
		portaineree.OperationDockerNetworkDisconnect:            true,
		portaineree.OperationDockerNetworkPrune:                 true,
		portaineree.OperationDockerNetworkDelete:                true,
		portaineree.OperationDockerVolumeList:                   true,
		portaineree.OperationDockerVolumeInspect:                true,
		portaineree.OperationDockerVolumeCreate:                 true,
		portaineree.OperationDockerVolumePrune:                  true,
		portaineree.OperationDockerVolumeDelete:                 true,
		portaineree.OperationDockerExecInspect:                  true,
		portaineree.OperationDockerExecStart:                    true,
		portaineree.OperationDockerExecResize:                   true,
		portaineree.OperationDockerSwarmInspect:                 true,
		portaineree.OperationDockerSwarmUnlockKey:               true,
		portaineree.OperationDockerSwarmInit:                    true,
		portaineree.OperationDockerSwarmJoin:                    true,
		portaineree.OperationDockerSwarmLeave:                   true,
		portaineree.OperationDockerSwarmUpdate:                  true,
		portaineree.OperationDockerSwarmUnlock:                  true,
		portaineree.OperationDockerNodeList:                     true,
		portaineree.OperationDockerNodeInspect:                  true,
		portaineree.OperationDockerNodeUpdate:                   true,
		portaineree.OperationDockerNodeDelete:                   true,
		portaineree.OperationDockerServiceList:                  true,
		portaineree.OperationDockerServiceInspect:               true,
		portaineree.OperationDockerServiceLogs:                  true,
		portaineree.OperationDockerServiceCreate:                true,
		portaineree.OperationDockerServiceUpdate:                true,
		portaineree.OperationDockerServiceForceUpdateService:    true,
		portaineree.OperationDockerServiceDelete:                true,
		portaineree.OperationDockerSecretList:                   true,
		portaineree.OperationDockerSecretInspect:                true,
		portaineree.OperationDockerSecretCreate:                 true,
		portaineree.OperationDockerSecretUpdate:                 true,
		portaineree.OperationDockerSecretDelete:                 true,
		portaineree.OperationDockerConfigList:                   true,
		portaineree.OperationDockerConfigInspect:                true,
		portaineree.OperationDockerConfigCreate:                 true,
		portaineree.OperationDockerConfigUpdate:                 true,
		portaineree.OperationDockerConfigDelete:                 true,
		portaineree.OperationDockerTaskList:                     true,
		portaineree.OperationDockerTaskInspect:                  true,
		portaineree.OperationDockerTaskLogs:                     true,
		portaineree.OperationDockerPluginList:                   true,
		portaineree.OperationDockerPluginPrivileges:             true,
		portaineree.OperationDockerPluginInspect:                true,
		portaineree.OperationDockerPluginPull:                   true,
		portaineree.OperationDockerPluginCreate:                 true,
		portaineree.OperationDockerPluginEnable:                 true,
		portaineree.OperationDockerPluginDisable:                true,
		portaineree.OperationDockerPluginPush:                   true,
		portaineree.OperationDockerPluginUpgrade:                true,
		portaineree.OperationDockerPluginSet:                    true,
		portaineree.OperationDockerPluginDelete:                 true,
		portaineree.OperationDockerSessionStart:                 true,
		portaineree.OperationDockerDistributionInspect:          true,
		portaineree.OperationDockerBuildPrune:                   true,
		portaineree.OperationDockerBuildCancel:                  true,
		portaineree.OperationDockerPing:                         true,
		portaineree.OperationDockerInfo:                         true,
		portaineree.OperationDockerVersion:                      true,
		portaineree.OperationDockerEvents:                       true,
		portaineree.OperationDockerSystem:                       true,
		portaineree.OperationDockerUndefined:                    true,
		portaineree.OperationDockerAgentPing:                    true,
		portaineree.OperationDockerAgentList:                    true,
		portaineree.OperationDockerAgentHostInfo:                true,
		portaineree.OperationDockerAgentBrowseDelete:            true,
		portaineree.OperationDockerAgentBrowseGet:               true,
		portaineree.OperationDockerAgentBrowseList:              true,
		portaineree.OperationDockerAgentBrowsePut:               true,
		portaineree.OperationDockerAgentBrowseRename:            true,
		portaineree.OperationDockerAgentUndefined:               true,
		portaineree.OperationPortainerResourceControlCreate:     true,
		portaineree.OperationPortainerResourceControlUpdate:     true,
		portaineree.OperationPortainerRegistryList:              true,
		portaineree.OperationPortainerRegistryInspect:           true,
		portaineree.OperationPortainerRegistryUpdateAccess:      true,
		portaineree.OperationPortainerStackList:                 true,
		portaineree.OperationPortainerStackInspect:              true,
		portaineree.OperationPortainerStackFile:                 true,
		portaineree.OperationPortainerStackCreate:               true,
		portaineree.OperationPortainerStackMigrate:              true,
		portaineree.OperationPortainerStackUpdate:               true,
		portaineree.OperationPortainerStackDelete:               true,
		portaineree.OperationPortainerWebsocketExec:             true,
		portaineree.OperationPortainerWebhookList:               true,
		portaineree.OperationPortainerWebhookCreate:             true,
		portaineree.OperationPortainerWebhookDelete:             true,
		portaineree.OperationPortainerEndpointUpdateSettings:    true,
		portaineree.OperationIntegrationStoridgeAdmin:           true,
		portaineree.EndpointResourcesAccess:                     true,
		portaineree.OperationHelmRepoList:                       true,
		portaineree.OperationHelmRepoCreate:                     true,
		portaineree.OperationHelmInstallChart:                   true,
		portaineree.OperationHelmUninstallChart:                 true,
	},
		DefaultK8sClusterAuthorizations()[portaineree.RoleIDEndpointAdmin],
		DefaultAzureAuthorizations()[portaineree.RoleIDEndpointAdmin],
	)
}

// DefaultEndpointAuthorizationsForHelpDeskRole returns the default environment(endpoint) authorizations
// associated to the helpdesk role.
func DefaultEndpointAuthorizationsForHelpDeskRole() portaineree.Authorizations {
	authorizations := unionAuthorizations(map[portaineree.Authorization]bool{
		portaineree.OperationDockerContainerArchiveInfo: true,
		portaineree.OperationDockerContainerList:        true,
		portaineree.OperationDockerContainerChanges:     true,
		portaineree.OperationDockerContainerInspect:     true,
		portaineree.OperationDockerContainerTop:         true,
		portaineree.OperationDockerContainerLogs:        true,
		portaineree.OperationDockerContainerStats:       true,
		portaineree.OperationDockerImageList:            true,
		portaineree.OperationDockerImageSearch:          true,
		portaineree.OperationDockerImageGetAll:          true,
		portaineree.OperationDockerImageGet:             true,
		portaineree.OperationDockerImageHistory:         true,
		portaineree.OperationDockerImageInspect:         true,
		portaineree.OperationDockerNetworkList:          true,
		portaineree.OperationDockerNetworkInspect:       true,
		portaineree.OperationDockerVolumeList:           true,
		portaineree.OperationDockerVolumeInspect:        true,
		portaineree.OperationDockerSwarmInspect:         true,
		portaineree.OperationDockerNodeList:             true,
		portaineree.OperationDockerNodeInspect:          true,
		portaineree.OperationDockerServiceList:          true,
		portaineree.OperationDockerServiceInspect:       true,
		portaineree.OperationDockerServiceLogs:          true,
		portaineree.OperationDockerSecretList:           true,
		portaineree.OperationDockerSecretInspect:        true,
		portaineree.OperationDockerConfigList:           true,
		portaineree.OperationDockerConfigInspect:        true,
		portaineree.OperationDockerTaskList:             true,
		portaineree.OperationDockerTaskInspect:          true,
		portaineree.OperationDockerTaskLogs:             true,
		portaineree.OperationDockerPluginList:           true,
		portaineree.OperationDockerDistributionInspect:  true,
		portaineree.OperationDockerPing:                 true,
		portaineree.OperationDockerInfo:                 true,
		portaineree.OperationDockerVersion:              true,
		portaineree.OperationDockerEvents:               true,
		portaineree.OperationDockerSystem:               true,
		portaineree.OperationDockerAgentPing:            true,
		portaineree.OperationDockerAgentList:            true,
		portaineree.OperationDockerAgentHostInfo:        true,
		portaineree.OperationPortainerStackList:         true,
		portaineree.OperationPortainerStackInspect:      true,
		portaineree.OperationPortainerStackFile:         true,
		portaineree.OperationPortainerWebhookList:       true,
		portaineree.EndpointResourcesAccess:             true,
	},
		DefaultK8sClusterAuthorizations()[portaineree.RoleIDHelpdesk],
		DefaultAzureAuthorizations()[portaineree.RoleIDHelpdesk],
	)

	return authorizations
}

// DefaultEndpointAuthorizationsForOperatorRole returns the default environment(endpoint) authorizations
// associated to the Operator role.
func DefaultEndpointAuthorizationsForOperatorRole() portaineree.Authorizations {
	authorizations := unionAuthorizations(map[portaineree.Authorization]bool{
		portaineree.OperationDockerContainerArchiveInfo:      true,
		portaineree.OperationDockerContainerList:             true,
		portaineree.OperationDockerContainerChanges:          true,
		portaineree.OperationDockerContainerInspect:          true,
		portaineree.OperationDockerContainerTop:              true,
		portaineree.OperationDockerContainerLogs:             true,
		portaineree.OperationDockerContainerStats:            true,
		portaineree.OperationDockerContainerKill:             true,
		portaineree.OperationDockerContainerPause:            true,
		portaineree.OperationDockerContainerUnpause:          true,
		portaineree.OperationDockerContainerRestart:          true,
		portaineree.OperationDockerContainerStart:            true,
		portaineree.OperationDockerContainerStop:             true,
		portaineree.OperationDockerContainerAttach:           true,
		portaineree.OperationDockerContainerExec:             true,
		portaineree.OperationDockerContainerResize:           true,
		portaineree.OperationDockerImageList:                 true,
		portaineree.OperationDockerImageSearch:               true,
		portaineree.OperationDockerImageGetAll:               true,
		portaineree.OperationDockerImageGet:                  true,
		portaineree.OperationDockerImageHistory:              true,
		portaineree.OperationDockerImageInspect:              true,
		portaineree.OperationDockerNetworkList:               true,
		portaineree.OperationDockerNetworkInspect:            true,
		portaineree.OperationDockerVolumeList:                true,
		portaineree.OperationDockerVolumeInspect:             true,
		portaineree.OperationDockerExecStart:                 true,
		portaineree.OperationDockerExecResize:                true,
		portaineree.OperationDockerSwarmInspect:              true,
		portaineree.OperationDockerNodeList:                  true,
		portaineree.OperationDockerNodeInspect:               true,
		portaineree.OperationDockerServiceList:               true,
		portaineree.OperationDockerServiceInspect:            true,
		portaineree.OperationDockerServiceLogs:               true,
		portaineree.OperationDockerServiceForceUpdateService: true,
		portaineree.OperationDockerSecretList:                true,
		portaineree.OperationDockerSecretInspect:             true,
		portaineree.OperationDockerConfigList:                true,
		portaineree.OperationDockerConfigInspect:             true,
		portaineree.OperationDockerTaskList:                  true,
		portaineree.OperationDockerTaskInspect:               true,
		portaineree.OperationDockerTaskLogs:                  true,
		portaineree.OperationDockerPluginList:                true,
		portaineree.OperationDockerDistributionInspect:       true,
		portaineree.OperationDockerPing:                      true,
		portaineree.OperationDockerInfo:                      true,
		portaineree.OperationDockerVersion:                   true,
		portaineree.OperationDockerEvents:                    true,
		portaineree.OperationDockerSystem:                    true,
		portaineree.OperationDockerAgentPing:                 true,
		portaineree.OperationDockerAgentList:                 true,
		portaineree.OperationDockerAgentHostInfo:             true,
		portaineree.OperationPortainerStackList:              true,
		portaineree.OperationPortainerStackInspect:           true,
		portaineree.OperationPortainerStackFile:              true,
		portaineree.OperationPortainerWebsocketExec:          true,
		portaineree.OperationPortainerWebhookList:            true,
		portaineree.EndpointResourcesAccess:                  true,
	},
		DefaultK8sClusterAuthorizations()[portaineree.RoleIDOperator],
		DefaultAzureAuthorizations()[portaineree.RoleIDOperator],
	)

	return authorizations
}

// DefaultEndpointAuthorizationsForStandardUserRole returns the default environment(endpoint) authorizations
// associated to the standard user role.
func DefaultEndpointAuthorizationsForStandardUserRole() portaineree.Authorizations {
	authorizations := unionAuthorizations(map[portaineree.Authorization]bool{
		portaineree.OperationDockerContainerArchiveInfo:         true,
		portaineree.OperationDockerContainerList:                true,
		portaineree.OperationDockerContainerExport:              true,
		portaineree.OperationDockerContainerChanges:             true,
		portaineree.OperationDockerContainerInspect:             true,
		portaineree.OperationDockerContainerTop:                 true,
		portaineree.OperationDockerContainerLogs:                true,
		portaineree.OperationDockerContainerStats:               true,
		portaineree.OperationDockerContainerAttachWebsocket:     true,
		portaineree.OperationDockerContainerArchive:             true,
		portaineree.OperationDockerContainerCreate:              true,
		portaineree.OperationDockerContainerKill:                true,
		portaineree.OperationDockerContainerPause:               true,
		portaineree.OperationDockerContainerUnpause:             true,
		portaineree.OperationDockerContainerRestart:             true,
		portaineree.OperationDockerContainerStart:               true,
		portaineree.OperationDockerContainerStop:                true,
		portaineree.OperationDockerContainerWait:                true,
		portaineree.OperationDockerContainerResize:              true,
		portaineree.OperationDockerContainerAttach:              true,
		portaineree.OperationDockerContainerExec:                true,
		portaineree.OperationDockerContainerRename:              true,
		portaineree.OperationDockerContainerUpdate:              true,
		portaineree.OperationDockerContainerPutContainerArchive: true,
		portaineree.OperationDockerContainerDelete:              true,
		portaineree.OperationDockerImageList:                    true,
		portaineree.OperationDockerImageSearch:                  true,
		portaineree.OperationDockerImageGetAll:                  true,
		portaineree.OperationDockerImageGet:                     true,
		portaineree.OperationDockerImageHistory:                 true,
		portaineree.OperationDockerImageInspect:                 true,
		portaineree.OperationDockerImageLoad:                    true,
		portaineree.OperationDockerImageCreate:                  true,
		portaineree.OperationDockerImagePush:                    true,
		portaineree.OperationDockerImageTag:                     true,
		portaineree.OperationDockerImageDelete:                  true,
		portaineree.OperationDockerImageCommit:                  true,
		portaineree.OperationDockerImageBuild:                   true,
		portaineree.OperationDockerNetworkList:                  true,
		portaineree.OperationDockerNetworkInspect:               true,
		portaineree.OperationDockerNetworkCreate:                true,
		portaineree.OperationDockerNetworkConnect:               true,
		portaineree.OperationDockerNetworkDisconnect:            true,
		portaineree.OperationDockerNetworkDelete:                true,
		portaineree.OperationDockerVolumeList:                   true,
		portaineree.OperationDockerVolumeInspect:                true,
		portaineree.OperationDockerVolumeCreate:                 true,
		portaineree.OperationDockerVolumeDelete:                 true,
		portaineree.OperationDockerExecInspect:                  true,
		portaineree.OperationDockerExecStart:                    true,
		portaineree.OperationDockerExecResize:                   true,
		portaineree.OperationDockerSwarmInspect:                 true,
		portaineree.OperationDockerSwarmUnlockKey:               true,
		portaineree.OperationDockerSwarmInit:                    true,
		portaineree.OperationDockerSwarmJoin:                    true,
		portaineree.OperationDockerSwarmLeave:                   true,
		portaineree.OperationDockerSwarmUpdate:                  true,
		portaineree.OperationDockerSwarmUnlock:                  true,
		portaineree.OperationDockerNodeList:                     true,
		portaineree.OperationDockerNodeInspect:                  true,
		portaineree.OperationDockerNodeUpdate:                   true,
		portaineree.OperationDockerNodeDelete:                   true,
		portaineree.OperationDockerServiceList:                  true,
		portaineree.OperationDockerServiceInspect:               true,
		portaineree.OperationDockerServiceLogs:                  true,
		portaineree.OperationDockerServiceCreate:                true,
		portaineree.OperationDockerServiceUpdate:                true,
		portaineree.OperationDockerServiceForceUpdateService:    true,
		portaineree.OperationDockerServiceDelete:                true,
		portaineree.OperationDockerSecretList:                   true,
		portaineree.OperationDockerSecretInspect:                true,
		portaineree.OperationDockerSecretCreate:                 true,
		portaineree.OperationDockerSecretUpdate:                 true,
		portaineree.OperationDockerSecretDelete:                 true,
		portaineree.OperationDockerConfigList:                   true,
		portaineree.OperationDockerConfigInspect:                true,
		portaineree.OperationDockerConfigCreate:                 true,
		portaineree.OperationDockerConfigUpdate:                 true,
		portaineree.OperationDockerConfigDelete:                 true,
		portaineree.OperationDockerTaskList:                     true,
		portaineree.OperationDockerTaskInspect:                  true,
		portaineree.OperationDockerTaskLogs:                     true,
		portaineree.OperationDockerPluginList:                   true,
		portaineree.OperationDockerPluginPrivileges:             true,
		portaineree.OperationDockerPluginInspect:                true,
		portaineree.OperationDockerPluginPull:                   true,
		portaineree.OperationDockerPluginCreate:                 true,
		portaineree.OperationDockerPluginEnable:                 true,
		portaineree.OperationDockerPluginDisable:                true,
		portaineree.OperationDockerPluginPush:                   true,
		portaineree.OperationDockerPluginUpgrade:                true,
		portaineree.OperationDockerPluginSet:                    true,
		portaineree.OperationDockerPluginDelete:                 true,
		portaineree.OperationDockerSessionStart:                 true,
		portaineree.OperationDockerDistributionInspect:          true,
		portaineree.OperationDockerBuildPrune:                   true,
		portaineree.OperationDockerBuildCancel:                  true,
		portaineree.OperationDockerPing:                         true,
		portaineree.OperationDockerInfo:                         true,
		portaineree.OperationDockerVersion:                      true,
		portaineree.OperationDockerEvents:                       true,
		portaineree.OperationDockerSystem:                       true,
		portaineree.OperationDockerUndefined:                    true,
		portaineree.OperationDockerAgentPing:                    true,
		portaineree.OperationDockerAgentList:                    true,
		portaineree.OperationDockerAgentHostInfo:                true,
		portaineree.OperationDockerAgentUndefined:               true,
		portaineree.OperationPortainerDockerHubInspect:          true,
		portaineree.OperationPortainerResourceControlUpdate:     true,
		portaineree.OperationPortainerStackList:                 true,
		portaineree.OperationPortainerStackInspect:              true,
		portaineree.OperationPortainerStackFile:                 true,
		portaineree.OperationPortainerStackCreate:               true,
		portaineree.OperationPortainerStackMigrate:              true,
		portaineree.OperationPortainerStackUpdate:               true,
		portaineree.OperationPortainerStackDelete:               true,
		portaineree.OperationPortainerWebsocketExec:             true,
		portaineree.OperationPortainerWebhookList:               true,
		portaineree.OperationPortainerWebhookCreate:             true,
		portaineree.OperationHelmRepoList:                       true,
		portaineree.OperationHelmRepoCreate:                     true,
		portaineree.OperationHelmInstallChart:                   true,
		portaineree.OperationHelmUninstallChart:                 true,
	},
		DefaultK8sClusterAuthorizations()[portaineree.RoleIDStandardUser],
		DefaultAzureAuthorizations()[portaineree.RoleIDStandardUser],
	)

	return authorizations
}

// DefaultEndpointAuthorizationsForReadOnlyUserRole returns the default environment(endpoint) authorizations
// associated to the readonly user role.
func DefaultEndpointAuthorizationsForReadOnlyUserRole() portaineree.Authorizations {
	authorizations := unionAuthorizations(map[portaineree.Authorization]bool{
		portaineree.OperationDockerContainerArchiveInfo: true,
		portaineree.OperationDockerContainerList:        true,
		portaineree.OperationDockerContainerChanges:     true,
		portaineree.OperationDockerContainerInspect:     true,
		portaineree.OperationDockerContainerTop:         true,
		portaineree.OperationDockerContainerLogs:        true,
		portaineree.OperationDockerContainerStats:       true,
		portaineree.OperationDockerImageList:            true,
		portaineree.OperationDockerImageSearch:          true,
		portaineree.OperationDockerImageGetAll:          true,
		portaineree.OperationDockerImageGet:             true,
		portaineree.OperationDockerImageHistory:         true,
		portaineree.OperationDockerImageInspect:         true,
		portaineree.OperationDockerNetworkList:          true,
		portaineree.OperationDockerNetworkInspect:       true,
		portaineree.OperationDockerVolumeList:           true,
		portaineree.OperationDockerVolumeInspect:        true,
		portaineree.OperationDockerSwarmInspect:         true,
		portaineree.OperationDockerNodeList:             true,
		portaineree.OperationDockerNodeInspect:          true,
		portaineree.OperationDockerServiceList:          true,
		portaineree.OperationDockerServiceInspect:       true,
		portaineree.OperationDockerServiceLogs:          true,
		portaineree.OperationDockerSecretList:           true,
		portaineree.OperationDockerSecretInspect:        true,
		portaineree.OperationDockerConfigList:           true,
		portaineree.OperationDockerConfigInspect:        true,
		portaineree.OperationDockerTaskList:             true,
		portaineree.OperationDockerTaskInspect:          true,
		portaineree.OperationDockerTaskLogs:             true,
		portaineree.OperationDockerPluginList:           true,
		portaineree.OperationDockerDistributionInspect:  true,
		portaineree.OperationDockerPing:                 true,
		portaineree.OperationDockerInfo:                 true,
		portaineree.OperationDockerVersion:              true,
		portaineree.OperationDockerEvents:               true,
		portaineree.OperationDockerSystem:               true,
		portaineree.OperationDockerAgentPing:            true,
		portaineree.OperationDockerAgentList:            true,
		portaineree.OperationDockerAgentHostInfo:        true,
		portaineree.OperationPortainerStackList:         true,
		portaineree.OperationPortainerStackInspect:      true,
		portaineree.OperationPortainerStackFile:         true,
		portaineree.OperationPortainerWebhookList:       true,
	},
		DefaultK8sClusterAuthorizations()[portaineree.RoleIDReadonly],
		DefaultAzureAuthorizations()[portaineree.RoleIDReadonly],
	)

	return authorizations
}

// DefaultPortainerAuthorizations returns the default Portainer authorizations used by non-admin users.
func DefaultPortainerAuthorizations() portaineree.Authorizations {
	return map[portaineree.Authorization]bool{
		portaineree.OperationPortainerEndpointGroupInspect:    true,
		portaineree.OperationPortainerEndpointGroupList:       true,
		portaineree.OperationPortainerDockerHubInspect:        true,
		portaineree.OperationPortainerEndpointList:            true,
		portaineree.OperationPortainerEndpointInspect:         true,
		portaineree.OperationPortainerEndpointExtensionAdd:    true,
		portaineree.OperationPortainerEndpointExtensionRemove: true,
		portaineree.OperationPortainerMOTD:                    true,
		portaineree.OperationPortainerRoleList:                true,
		portaineree.OperationPortainerTeamList:                true,
		portaineree.OperationPortainerTemplateList:            true,
		portaineree.OperationPortainerTemplateInspect:         true,
		portaineree.OperationPortainerUserList:                true,
		portaineree.OperationPortainerUserInspect:             true,
		portaineree.OperationPortainerUserMemberships:         true,
		portaineree.OperationPortainerUserListToken:           true,
		portaineree.OperationPortainerUserCreateToken:         true,
		portaineree.OperationPortainerUserRevokeToken:         true,
	}
}

// RegisterEventHandler upserts event handler by id
func (service *Service) RegisterEventHandler(id string, handler portaineree.AuthEventHandler) {
	service.authEventHandlers[id] = handler
}

// TriggerEndpointAuthUpdate triggers environment(endpoint) auth update event on the registered
// event handlers (e.g. token cache manager)
func (service *Service) TriggerEndpointAuthUpdate(endpointID int) {
	for _, handler := range service.authEventHandlers {
		handler.HandleEndpointAuthUpdate(endpointID)
	}
}

// TriggerUserAuthUpdate triggers all users auth update event on the registered
// event handlers (e.g. token cache manager)
func (service *Service) TriggerUsersAuthUpdate() {
	for _, handler := range service.authEventHandlers {
		handler.HandleUsersAuthUpdate()
	}
}

// TriggerUserAuthUpdate triggers single user auth update event on the registered
// event handlers (e.g. token cache manager)
func (service *Service) TriggerUserAuthUpdate(userID int) {
	for _, handler := range service.authEventHandlers {
		handler.HandleUserAuthDelete(userID)
	}
}

func populateVolumeBrowsingAuthorizations(rolePointer *portaineree.Role) portaineree.Role {
	role := *rolePointer

	role.Authorizations[portaineree.OperationDockerAgentBrowseGet] = true
	role.Authorizations[portaineree.OperationDockerAgentBrowseList] = true

	if role.ID == portaineree.RoleIDStandardUser {
		role.Authorizations[portaineree.OperationDockerAgentBrowseDelete] = true
		role.Authorizations[portaineree.OperationDockerAgentBrowsePut] = true
		role.Authorizations[portaineree.OperationDockerAgentBrowseRename] = true
	}

	return role
}

// RemoveTeamAccessPolicies will remove all existing access policies associated to the specified team
func (service *Service) RemoveTeamAccessPolicies(teamID portaineree.TeamID) error {
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		for policyTeamID := range endpoint.TeamAccessPolicies {
			if policyTeamID == teamID {
				delete(endpoint.TeamAccessPolicies, policyTeamID)

				err := service.dataStore.Endpoint().UpdateEndpoint(endpoint.ID, &endpoint)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	endpointGroups, err := service.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return err
	}

	for _, endpointGroup := range endpointGroups {
		for policyTeamID := range endpointGroup.TeamAccessPolicies {
			if policyTeamID == teamID {
				delete(endpointGroup.TeamAccessPolicies, policyTeamID)

				err := service.dataStore.EndpointGroup().UpdateEndpointGroup(endpointGroup.ID, &endpointGroup)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	registries, err := service.dataStore.Registry().Registries()
	if err != nil {
		return err
	}

	for _, registry := range registries {
		for policyTeamID := range registry.TeamAccessPolicies {
			if policyTeamID == teamID {
				delete(registry.TeamAccessPolicies, policyTeamID)

				err := service.dataStore.Registry().UpdateRegistry(registry.ID, &registry)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	return service.UpdateUsersAuthorizations()
}

// RemoveUserAccessPolicies will remove all existing access policies associated to the specified user
func (service *Service) RemoveUserAccessPolicies(userID portaineree.UserID) error {
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		for policyUserID := range endpoint.UserAccessPolicies {
			if policyUserID == userID {
				delete(endpoint.UserAccessPolicies, policyUserID)

				err := service.dataStore.Endpoint().UpdateEndpoint(endpoint.ID, &endpoint)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	endpointGroups, err := service.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return err
	}

	for _, endpointGroup := range endpointGroups {
		for policyUserID := range endpointGroup.UserAccessPolicies {
			if policyUserID == userID {
				delete(endpointGroup.UserAccessPolicies, policyUserID)

				err := service.dataStore.EndpointGroup().UpdateEndpointGroup(endpointGroup.ID, &endpointGroup)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	registries, err := service.dataStore.Registry().Registries()
	if err != nil {
		return err
	}

	for _, registry := range registries {
		for policyUserID := range registry.UserAccessPolicies {
			if policyUserID == userID {
				delete(registry.UserAccessPolicies, policyUserID)

				err := service.dataStore.Registry().UpdateRegistry(registry.ID, &registry)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	service.TriggerUserAuthUpdate(int(userID))

	return nil
}

// UpdateUsersAuthorizations will trigger an update of the authorizations for all the users.
func (service *Service) UpdateUsersAuthorizations() error {
	users, err := service.dataStore.User().Users()
	if err != nil {
		return err
	}

	for _, user := range users {
		err := service.updateUserAuthorizations(user.ID)
		if err != nil {
			return err
		}
	}

	service.TriggerUsersAuthUpdate()

	return nil
}

func (service *Service) updateUserAuthorizations(userID portaineree.UserID) error {
	user, err := service.dataStore.User().User(userID)
	if err != nil {
		return err
	}

	endpointAuthorizations, err := service.getAuthorizations(user)
	if err != nil {
		return err
	}

	user.EndpointAuthorizations = endpointAuthorizations

	return service.dataStore.User().UpdateUser(userID, user)
}

func (service *Service) getAuthorizations(user *portaineree.User) (portaineree.EndpointAuthorizations, error) {
	endpointAuthorizations := portaineree.EndpointAuthorizations{}
	if user.Role == portaineree.AdministratorRole {
		return endpointAuthorizations, nil
	}

	userMemberships, err := service.dataStore.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return endpointAuthorizations, err
	}

	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		return endpointAuthorizations, err
	}

	endpointGroups, err := service.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return endpointAuthorizations, err
	}

	roles, err := service.dataStore.Role().Roles()
	if err != nil {
		return endpointAuthorizations, err
	}

	endpointAuthorizations = getUserEndpointAuthorizations(user, endpoints, endpointGroups, roles, userMemberships)

	return endpointAuthorizations, nil
}

func getUserEndpointAuthorizations(user *portaineree.User, endpoints []portaineree.Endpoint,
	endpointGroups []portaineree.EndpointGroup, roles []portaineree.Role,
	userMemberships []portaineree.TeamMembership) portaineree.EndpointAuthorizations {

	endpointAuthorizations := make(portaineree.EndpointAuthorizations)
	for endpointID, role := range getUserEndpointRoles(user, endpoints,
		endpointGroups, roles, userMemberships) {
		endpointAuthorizations[endpointID] = role.Authorizations
	}

	return endpointAuthorizations
}

// get the user and team policies from the environment(endpoint) group definitions
func getGroupPolicies(endpointGroups []portaineree.EndpointGroup) (
	map[portaineree.EndpointGroupID]portaineree.UserAccessPolicies,
	map[portaineree.EndpointGroupID]portaineree.TeamAccessPolicies,
) {
	groupUserAccessPolicies := map[portaineree.EndpointGroupID]portaineree.UserAccessPolicies{}
	groupTeamAccessPolicies := map[portaineree.EndpointGroupID]portaineree.TeamAccessPolicies{}
	for _, endpointGroup := range endpointGroups {
		groupUserAccessPolicies[endpointGroup.ID] = endpointGroup.UserAccessPolicies
		groupTeamAccessPolicies[endpointGroup.ID] = endpointGroup.TeamAccessPolicies
	}
	return groupUserAccessPolicies, groupTeamAccessPolicies
}

// UpdateUserNamespaceAccessPolicies takes an input accessPolicies
// and updates it with the user and his team's environment(endpoint) roles.
// Returns the updated policies and whether there is any update.
func (service *Service) UpdateUserNamespaceAccessPolicies(
	userID int, endpoint *portaineree.Endpoint,
	policiesToUpdate map[string]portaineree.K8sNamespaceAccessPolicy,
) (map[string]portaineree.K8sNamespaceAccessPolicy, bool, error) {
	endpointID := int(endpoint.ID)
	restrictDefaultNamespace := endpoint.Kubernetes.Configuration.RestrictDefaultNamespace

	userRole, err := service.GetUserEndpointRole(userID, endpointID)
	if err != nil {
		return nil, false, err
	}
	usersEndpointRole := make(map[int]int)
	teamsEndpointRole := make(map[int]int)
	if userRole != nil {
		usersEndpointRole[userID] = int(userRole.ID)
	} else {
		usersEndpointRole[userID] = -1
	}

	userMemberships, err := service.dataStore.TeamMembership().
		TeamMembershipsByUserID(portaineree.UserID(userID))
	if err != nil {
		return nil, false, err
	}
	teamIDs := make([]int, 0)
	for _, membership := range userMemberships {
		teamRole, err := service.GetTeamEndpointRole(int(membership.TeamID), endpointID)
		if err != nil {
			return nil, false, err
		}
		if teamRole != nil {
			teamsEndpointRole[int(membership.TeamID)] = int(teamRole.ID)
			teamIDs = append(teamIDs, int(membership.TeamID))
		}
	}
	return service.updateNamespaceAccessPolicies(userID, teamIDs, usersEndpointRole, teamsEndpointRole,
		policiesToUpdate, restrictDefaultNamespace)
}

// updateNamespaceAccessPolicies takes an input accessPolicies
// and updates it with the environment(endpoint) users/teams roles.
func (service *Service) updateNamespaceAccessPolicies(
	selectedUserID int, selectedTeamIDs []int,
	usersEndpointRole map[int]int, teamsEndpointRole map[int]int,
	policiesToUpdate map[string]portaineree.K8sNamespaceAccessPolicy,
	restrictDefaultNamespace bool,
) (map[string]portaineree.K8sNamespaceAccessPolicy, bool, error) {
	hasChange := false
	if !restrictDefaultNamespace {
		delete(policiesToUpdate, "default")
		hasChange = true
	}
	for ns, nsPolicies := range policiesToUpdate {
		for userID, policy := range nsPolicies.UserAccessPolicies {
			if int(userID) == selectedUserID {
				iRoleID, ok := usersEndpointRole[int(userID)]
				if !ok {
					delete(nsPolicies.UserAccessPolicies, userID)
					hasChange = true
				} else if int(policy.RoleID) != iRoleID {
					nsPolicies.UserAccessPolicies[userID] = portaineree.AccessPolicy{
						RoleID: portaineree.RoleID(iRoleID),
					}
					hasChange = true
				}
			}
		}
		for teamID, policy := range nsPolicies.TeamAccessPolicies {
			for _, selectedTeamID := range selectedTeamIDs {
				if int(teamID) == selectedTeamID {
					iRoleID, ok := teamsEndpointRole[int(teamID)]
					if !ok {
						delete(nsPolicies.TeamAccessPolicies, teamID)
						hasChange = true
					} else if int(policy.RoleID) != iRoleID {
						nsPolicies.TeamAccessPolicies[teamID] = portaineree.AccessPolicy{
							RoleID: portaineree.RoleID(iRoleID),
						}
						hasChange = true
					}
				}
			}
		}
		policiesToUpdate[ns] = nsPolicies
	}
	return policiesToUpdate, hasChange, nil
}

// RemoveUserNamespaceAccessPolicies takes an input accessPolicies
// and remove users/teams in it.
// Returns the updated policies and whether there is any update.
func (service *Service) RemoveUserNamespaceAccessPolicies(
	userID int, endpointID int,
	policiesToUpdate map[string]portaineree.K8sNamespaceAccessPolicy,
) (map[string]portaineree.K8sNamespaceAccessPolicy, bool, error) {
	userRole, err := service.GetUserEndpointRole(userID, endpointID)
	if err != nil {
		return nil, false, err
	}
	usersEndpointRole := make(map[int]int)
	if userRole != nil {
		usersEndpointRole[userID] = int(userRole.ID)
	}
	return service.removeUserInNamespaceAccessPolicies(usersEndpointRole, policiesToUpdate)
}

// removeUserInNamespaceAccessPolicies takes an input accessPolicies
// and remove users/teams in it.
func (service *Service) removeUserInNamespaceAccessPolicies(
	usersEndpointRole map[int]int,
	policiesToUpdate map[string]portaineree.K8sNamespaceAccessPolicy,
) (map[string]portaineree.K8sNamespaceAccessPolicy, bool, error) {
	hasChange := false
	for ns, nsPolicies := range policiesToUpdate {
		for userID := range nsPolicies.UserAccessPolicies {
			_, ok := usersEndpointRole[int(userID)]
			if ok {
				delete(nsPolicies.UserAccessPolicies, userID)
				hasChange = true
			}
		}
		if len(nsPolicies.UserAccessPolicies) == 0 && len(nsPolicies.TeamAccessPolicies) == 0 {
			delete(policiesToUpdate, ns)
		} else {
			policiesToUpdate[ns] = nsPolicies
		}
	}
	return policiesToUpdate, hasChange, nil
}

// RemoveTeamsNamespaceAccessPolicies takes an input accessPolicies
// and remove teams in it.
// Returns the updated policies and whether there is any update.
func (service *Service) RemoveTeamNamespaceAccessPolicies(
	teamID int, endpointID int,
	policiesToUpdate map[string]portaineree.K8sNamespaceAccessPolicy,
) (map[string]portaineree.K8sNamespaceAccessPolicy, bool, error) {
	teamRole, err := service.GetTeamEndpointRole(teamID, endpointID)
	if err != nil {
		return nil, false, err
	}
	if teamRole == nil {
		return nil, false, nil
	}
	teamsEndpointRole := make(map[int]int)
	teamsEndpointRole[teamID] = int(teamRole.ID)

	hasChange := false
	for ns, nsPolicies := range policiesToUpdate {
		for teamID := range nsPolicies.TeamAccessPolicies {
			_, ok := teamsEndpointRole[int(teamID)]
			if ok {
				delete(nsPolicies.TeamAccessPolicies, teamID)
				hasChange = true
			}
		}
		if len(nsPolicies.UserAccessPolicies) == 0 && len(nsPolicies.TeamAccessPolicies) == 0 {
			delete(policiesToUpdate, ns)
		} else {
			policiesToUpdate[ns] = nsPolicies
		}
	}
	return policiesToUpdate, hasChange, nil
}

// GetUserEndpointRole returns the environment(endpoint) role of the user.
// It returns nil if there is no role assigned to the user at the environment(endpoint).
func (service *Service) GetUserEndpointRole(
	userID int,
	endpointID int,
) (*portaineree.Role, error) {
	user, err := service.dataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		return nil, err
	}

	userMemberships, err := service.dataStore.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	endpoint, err := service.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err != nil {
		return nil, err
	}

	endpointGroups, err := service.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return nil, err
	}

	roles, err := service.dataStore.Role().Roles()
	if err != nil {
		return nil, err
	}

	groupUserAccessPolicies, groupTeamAccessPolicies := getGroupPolicies(endpointGroups)

	return getUserEndpointRole(user, *endpoint, groupUserAccessPolicies,
		groupTeamAccessPolicies, roles, userMemberships), nil
}

func (service *Service) GetNamespaceAuthorizations(
	userID int,
	endpoint portaineree.Endpoint,
	kcl portaineree.KubeClient,
) (map[string]portaineree.Authorizations, error) {
	namespaceAuthorizations := make(map[string]portaineree.Authorizations)

	// skip non k8s environments(endpoints)
	if endpoint.Type != portaineree.KubernetesLocalEnvironment &&
		endpoint.Type != portaineree.AgentOnKubernetesEnvironment &&
		endpoint.Type != portaineree.EdgeAgentOnKubernetesEnvironment {
		return namespaceAuthorizations, nil
	}

	endpointRole, err := service.GetUserEndpointRole(userID, int(endpoint.ID))
	if err != nil {
		return nil, err
	}

	// no environment(endpoint) role for the user, continue
	if endpointRole == nil {
		return namespaceAuthorizations, nil
	}

	namespaces, err := kcl.GetNamespaces()
	if err != nil {
		return nil, err
	}

	accessPolicies, err := kcl.GetNamespaceAccessPolicies()
	if err != nil {
		return nil, err
	}
	// update the namespace access policies based on user's role, also in configmap.
	accessPolicies, hasChange, err := service.UpdateUserNamespaceAccessPolicies(
		userID, &endpoint, accessPolicies,
	)
	if hasChange {
		err = kcl.UpdateNamespaceAccessPolicies(accessPolicies)
		if err != nil {
			return nil, err
		}
	}

	namespaceAuthorizations, err = service.GetUserNamespaceAuthorizations(
		userID, int(endpointRole.ID), int(endpoint.ID), accessPolicies, namespaces, endpointRole.Authorizations,
		endpoint.Kubernetes.Configuration,
	)
	if err != nil {
		return nil, err
	}

	return namespaceAuthorizations, nil
}

// GetUserNamespaceAuthorizations returns authorizations of a user's namespaces
func (service *Service) GetUserNamespaceAuthorizations(
	userID int,
	userEndpointRoleID int,
	endpointID int,
	accessPolicies map[string]portaineree.K8sNamespaceAccessPolicy,
	namespaces map[string]portaineree.K8sNamespaceInfo,
	endpointAuthorizations portaineree.Authorizations,
	endpointConfiguration portaineree.KubernetesConfiguration,
) (map[string]portaineree.Authorizations, error) {
	namespaceRoles, err := service.GetUserNamespaceRoles(userID, userEndpointRoleID, endpointID,
		accessPolicies, namespaces, endpointAuthorizations, endpointConfiguration)
	if err != nil {
		return nil, err
	}

	defaultAuthorizations := DefaultK8sNamespaceAuthorizations()

	namespaceAuthorizations := make(map[string]portaineree.Authorizations)
	for namespace, role := range namespaceRoles {
		namespaceAuthorizations[namespace] = defaultAuthorizations[role.ID]
	}

	return namespaceAuthorizations, nil
}

// GetUserNamespaceRoles returns the environment(endpoint) role of the user.
func (service *Service) GetUserNamespaceRoles(
	userID int,
	userEndpointRoleID int,
	endpointID int,
	accessPolicies map[string]portaineree.K8sNamespaceAccessPolicy,
	namespaces map[string]portaineree.K8sNamespaceInfo,
	endpointAuthorizations portaineree.Authorizations,
	endpointConfiguration portaineree.KubernetesConfiguration,
) (map[string]portaineree.Role, error) {

	// does an early check if user can access all namespaces to skip db calls
	accessAllNamespaces := endpointAuthorizations[portaineree.OperationK8sAccessAllNamespaces]
	if accessAllNamespaces {
		return make(map[string]portaineree.Role), nil
	}

	user, err := service.dataStore.User().User(portaineree.UserID(userID))
	if err != nil {
		return nil, err
	}

	userMemberships, err := service.dataStore.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	roles, err := service.dataStore.Role().Roles()
	if err != nil {
		return nil, err
	}

	accessSystemNamespaces := endpointAuthorizations[portaineree.OperationK8sAccessSystemNamespaces]
	accessUserNamespaces := endpointAuthorizations[portaineree.OperationK8sAccessUserNamespaces]

	return getUserNamespaceRoles(user, userEndpointRoleID, roles, userMemberships,
		accessPolicies, namespaces, accessAllNamespaces, accessSystemNamespaces,
		accessUserNamespaces, endpointConfiguration.RestrictDefaultNamespace)
}

func getUserNamespaceRoles(
	user *portaineree.User,
	userEndpointRoleID int,
	roles []portaineree.Role,
	userMemberships []portaineree.TeamMembership,
	accessPolicies map[string]portaineree.K8sNamespaceAccessPolicy,
	namespaces map[string]portaineree.K8sNamespaceInfo,
	accessAllNamespaces bool,
	accessSystemNamespaces bool,
	accessUserNamespaces bool,
	restrictDefaultNamespace bool,
) (map[string]portaineree.Role, error) {
	rolesMap := make(map[int]portaineree.Role)
	for _, role := range roles {
		rolesMap[int(role.ID)] = role
	}
	results := make(map[string]portaineree.Role)

	for namespace, info := range namespaces {
		// user can access everything
		if accessAllNamespaces {
			results[namespace] = rolesMap[userEndpointRoleID]
		}

		// skip default namespace or system namespace (when user don't have access)
		if !accessSystemNamespaces && info.IsSystem {
			continue
		}

		// default namespace doesn't allow permission management so no role
		// aggregation needed
		if !restrictDefaultNamespace && info.IsDefault {
			results[namespace] = rolesMap[userEndpointRoleID]
		}

		// user can access user namespaces
		if accessUserNamespaces && !info.IsSystem && !info.IsDefault {
			results[namespace] = rolesMap[userEndpointRoleID]
		}

		// if there is an access policy for the current namespace
		if policies, ok := accessPolicies[namespace]; ok {
			role := getUserNamespaceRole(
				user,
				policies.UserAccessPolicies,
				policies.TeamAccessPolicies,
				roles,
				userMemberships,
			)
			if role != nil {
				results[namespace] = *role
			}
		}
	}

	return results, nil
}

// For each namespace, first calculate the role(s) of a user
// based on the sequence of searching:
//  - His namespace role (single)
//  - His teams namespace role (multiple, 1 user has n teams)
//
// If roles are found in any of the step, the search stops.
// Then the role with the hightest priority is returned.
func getUserNamespaceRole(
	user *portaineree.User,
	userAccessPolicies portaineree.UserAccessPolicies,
	teamAccessPolicies portaineree.TeamAccessPolicies,
	roles []portaineree.Role,
	userMemberships []portaineree.TeamMembership,
) *portaineree.Role {

	role := getRoleFromUserAccessPolicies(user, userAccessPolicies, roles)
	if role != nil {
		return role
	}

	role = getRoleFromTeamAccessPolicies(userMemberships, teamAccessPolicies, roles)
	return role
}

// For each environment(endpoint), first calculate the role(s) of a team
// based on the sequence of searching:
//  - Team's environment(endpoint) role (multiple, 1 user has n teams)
//  - Team's roles in all the assigned environment(endpoint) groups (multiple, 1 user has n teams, 1 team has 1 environment(endpoint) group)
//
// If roles are found in any of the step, the search stops.
// Then the role with the hightest priority is returned.
func (service *Service) GetTeamEndpointRole(
	teamID int, endpointID int,
) (*portaineree.Role, error) {

	memberships, err := service.dataStore.TeamMembership().TeamMembershipsByTeamID(portaineree.TeamID(teamID))
	if err != nil {
		return nil, err
	}

	endpoint, err := service.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err != nil {
		return nil, err
	}

	endpointGroups, err := service.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return nil, err
	}

	roles, err := service.dataStore.Role().Roles()
	if err != nil {
		return nil, err
	}

	_, groupTeamAccessPolicies := getGroupPolicies(endpointGroups)

	role := getRoleFromTeamAccessPolicies(memberships,
		endpoint.TeamAccessPolicies, roles)
	if role != nil {
		return role, nil
	}

	role = getRoleFromTeamEndpointGroupPolicies(memberships, endpoint,
		roles, groupTeamAccessPolicies)
	return role, nil
}

// For each environment(endpoint), first calculate the role(s) of a user
// based on the sequence of searching:
//  - His environment(endpoint) role (single)
//  - His environment(endpoint) group role (single, 1 environment(endpoint) has 1 environment(endpoint) group)
//  - His teams environment(endpoint) role (multiple, 1 user has n teams)
//  - His teams roles in all the assigned environment(endpoint) groups (multiple, 1 user has n teams, 1 team has 1 environment(endpoint) group)
//
// If roles are found in any of the step, the search stops.
// Then the role with the hightest priority is returned.
func getUserEndpointRole(user *portaineree.User, endpoint portaineree.Endpoint,
	groupUserAccessPolicies map[portaineree.EndpointGroupID]portaineree.UserAccessPolicies,
	groupTeamAccessPolicies map[portaineree.EndpointGroupID]portaineree.TeamAccessPolicies,
	roles []portaineree.Role,
	userMemberships []portaineree.TeamMembership,
) *portaineree.Role {

	role := getRoleFromUserAccessPolicies(user, endpoint.UserAccessPolicies, roles)
	if role == nil {
		role = getRoleFromUserEndpointGroupPolicy(user, &endpoint, roles, groupUserAccessPolicies)
	}
	if role == nil {
		role = getRoleFromTeamAccessPolicies(userMemberships, endpoint.TeamAccessPolicies, roles)
	}
	if role == nil {
		role = getRoleFromTeamEndpointGroupPolicies(userMemberships, &endpoint, roles, groupTeamAccessPolicies)
	}

	if role != nil && endpoint.SecuritySettings.AllowVolumeBrowserForRegularUsers {
		newRole := populateVolumeBrowsingAuthorizations(role)
		role = &newRole
	}

	return role
}

func getUserEndpointRoles(user *portaineree.User, endpoints []portaineree.Endpoint,
	endpointGroups []portaineree.EndpointGroup, roles []portaineree.Role,
	userMemberships []portaineree.TeamMembership) map[portaineree.EndpointID]portaineree.Role {
	results := make(map[portaineree.EndpointID]portaineree.Role)

	groupUserAccessPolicies, groupTeamAccessPolicies := getGroupPolicies(endpointGroups)

	for _, endpoint := range endpoints {
		role := getUserEndpointRole(user, endpoint, groupUserAccessPolicies,
			groupTeamAccessPolicies, roles, userMemberships)
		if role != nil {
			results[endpoint.ID] = *role
			continue
		}
	}

	return results
}

// A user may have 1 role in each assigned environments(endpoints).
func getRoleFromUserAccessPolicies(
	user *portaineree.User,
	userAccessPolicies portaineree.UserAccessPolicies,
	roles []portaineree.Role,
) *portaineree.Role {
	policyRoles := make([]portaineree.RoleID, 0)

	policy, ok := userAccessPolicies[user.ID]
	if ok {
		policyRoles = append(policyRoles, policy.RoleID)
	}
	if len(policyRoles) == 0 {
		return nil
	}

	return getKeyRole(policyRoles, roles)
}

// An environment(endpoint) can only have 1 EndpointGroup.
//
// A user may have 1 role in each assigned EndpointGroups.
func getRoleFromUserEndpointGroupPolicy(user *portaineree.User,
	endpoint *portaineree.Endpoint, roles []portaineree.Role,
	groupAccessPolicies map[portaineree.EndpointGroupID]portaineree.UserAccessPolicies) *portaineree.Role {
	policyRoles := make([]portaineree.RoleID, 0)

	policy, ok := groupAccessPolicies[endpoint.GroupID][user.ID]
	if ok {
		policyRoles = append(policyRoles, policy.RoleID)
	}
	if len(policyRoles) == 0 {
		return nil
	}

	return getKeyRole(policyRoles, roles)
}

// A team may have 1 role in each assigned environments(endpoints)
func getRoleFromTeamAccessPolicies(
	memberships []portaineree.TeamMembership,
	teamAccessPolicies portaineree.TeamAccessPolicies,
	roles []portaineree.Role,
) *portaineree.Role {
	policyRoles := make([]portaineree.RoleID, 0)

	for _, membership := range memberships {
		policy, ok := teamAccessPolicies[membership.TeamID]
		if ok {
			policyRoles = append(policyRoles, policy.RoleID)
		}
	}
	if len(policyRoles) == 0 {
		return nil
	}

	return getKeyRole(policyRoles, roles)
}

// An environment(endpoint) can only have 1 EndpointGroup.
//
// A team may have 1 role in each assigned EndpointGroups.
func getRoleFromTeamEndpointGroupPolicies(memberships []portaineree.TeamMembership,
	endpoint *portaineree.Endpoint, roles []portaineree.Role,
	groupTeamAccessPolicies map[portaineree.EndpointGroupID]portaineree.TeamAccessPolicies) *portaineree.Role {
	policyRoles := make([]portaineree.RoleID, 0)

	for _, membership := range memberships {
		policy, ok := groupTeamAccessPolicies[endpoint.GroupID][membership.TeamID]
		if ok {
			policyRoles = append(policyRoles, policy.RoleID)
		}
	}
	if len(policyRoles) == 0 {
		return nil
	}

	return getKeyRole(policyRoles, roles)
}

// for each role in the roleIdentifiers,
// find the highest priority role and returns its authorizations
func getAuthorizationsFromRoles(roleIdentifiers []portaineree.RoleID, roles []portaineree.Role) portaineree.Authorizations {
	keyRole := getKeyRole(roleIdentifiers, roles)

	if keyRole == nil {
		return portaineree.Authorizations{}
	}

	return keyRole.Authorizations
}

// for each role in the roleIdentifiers,
// find the highest priority role
func getKeyRole(roleIdentifiers []portaineree.RoleID, roles []portaineree.Role) *portaineree.Role {
	var associatedRoles []portaineree.Role

	for _, id := range roleIdentifiers {
		for _, role := range roles {
			if role.ID == id {
				associatedRoles = append(associatedRoles, role)
				break
			}
		}
	}

	var result portaineree.Role
	for _, role := range associatedRoles {
		if role.Priority > result.Priority {
			result = role
		}
	}

	return &result
}

// unionAuthorizations returns a union of all the input authorizations
// using the "or" operator.
func unionAuthorizations(auths ...portaineree.Authorizations) portaineree.Authorizations {
	authorizations := make(portaineree.Authorizations)

	for _, auth := range auths {
		for authKey, authVal := range auth {
			if val, ok := authorizations[authKey]; ok {
				authorizations[authKey] = val || authVal
			} else {
				authorizations[authKey] = authVal
			}
		}
	}

	return authorizations
}
