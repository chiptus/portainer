package authorization

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// DefaultK8sClusterAuthorizations returns a set of default k8s cluster-level authorizations
// based on user's role. The operations are supposed to be used in front-end.
func DefaultK8sClusterAuthorizations() map[portaineree.RoleID]portaineree.Authorizations {
	return map[portaineree.RoleID]portaineree.Authorizations{
		portaineree.RoleIDEndpointAdmin: {
			portaineree.OperationK8sAccessAllNamespaces:              true,
			portaineree.OperationK8sAccessSystemNamespaces:           true,
			portaineree.OperationK8sAccessUserNamespaces:             true,
			portaineree.OperationK8sResourcePoolsR:                   true,
			portaineree.OperationK8sResourcePoolsW:                   true,
			portaineree.OperationK8sResourcePoolDetailsR:             true,
			portaineree.OperationK8sResourcePoolDetailsW:             true,
			portaineree.OperationK8sResourcePoolsAccessManagementRW:  true,
			portaineree.OperationK8sApplicationsR:                    true,
			portaineree.OperationK8sApplicationsW:                    true,
			portaineree.OperationK8sApplicationDetailsR:              true,
			portaineree.OperationK8sApplicationDetailsW:              true,
			portaineree.OperationK8sPodDelete:                        true,
			portaineree.OperationK8sApplicationConsoleRW:             true,
			portaineree.OperationK8sApplicationsAdvancedDeploymentRW: true,
			portaineree.OperationK8sConfigurationsR:                  true,
			portaineree.OperationK8sConfigurationsW:                  true,
			portaineree.OperationK8sConfigurationsDetailsR:           true,
			portaineree.OperationK8sConfigurationsDetailsW:           true,
			portaineree.OperationK8sRegistrySecretList:               true,
			portaineree.OperationK8sRegistrySecretInspect:            true,
			portaineree.OperationK8sVolumesR:                         true,
			portaineree.OperationK8sVolumesW:                         true,
			portaineree.OperationK8sVolumeDetailsR:                   true,
			portaineree.OperationK8sVolumeDetailsW:                   true,
			portaineree.OperationK8sClusterR:                         true,
			portaineree.OperationK8sClusterW:                         true,
			portaineree.OperationK8sClusterNodeR:                     true,
			portaineree.OperationK8sClusterNodeW:                     true,
			portaineree.OperationK8sClusterSetupRW:                   true,
			portaineree.OperationK8sApplicationErrorDetailsR:         true,
			portaineree.OperationK8sStorageClassDisabledR:            true,
		},
		portaineree.RoleIDOperator: {
			portaineree.OperationK8sAccessUserNamespaces:     true,
			portaineree.OperationK8sResourcePoolsR:           true,
			portaineree.OperationK8sResourcePoolDetailsR:     true,
			portaineree.OperationK8sApplicationsR:            true,
			portaineree.OperationK8sApplicationDetailsR:      true,
			portaineree.OperationK8sPodDelete:                true,
			portaineree.OperationK8sApplicationConsoleRW:     true,
			portaineree.OperationK8sConfigurationsR:          true,
			portaineree.OperationK8sConfigurationsDetailsR:   true,
			portaineree.OperationK8sConfigurationsDetailsW:   true,
			portaineree.OperationK8sVolumesR:                 true,
			portaineree.OperationK8sVolumeDetailsR:           true,
			portaineree.OperationK8sClusterR:                 true,
			portaineree.OperationK8sClusterNodeR:             true,
			portaineree.OperationK8sApplicationErrorDetailsR: true,
			portaineree.OperationK8sStorageClassDisabledR:    true,
		},
		portaineree.RoleIDHelpdesk: {
			portaineree.OperationK8sAccessUserNamespaces:     true,
			portaineree.OperationK8sResourcePoolsR:           true,
			portaineree.OperationK8sResourcePoolDetailsR:     true,
			portaineree.OperationK8sApplicationsR:            true,
			portaineree.OperationK8sApplicationDetailsR:      true,
			portaineree.OperationK8sConfigurationsR:          true,
			portaineree.OperationK8sConfigurationsDetailsR:   true,
			portaineree.OperationK8sVolumesR:                 true,
			portaineree.OperationK8sVolumeDetailsR:           true,
			portaineree.OperationK8sClusterR:                 true,
			portaineree.OperationK8sClusterNodeR:             true,
			portaineree.OperationK8sApplicationErrorDetailsR: true,
			portaineree.OperationK8sStorageClassDisabledR:    true,
		},
		portaineree.RoleIDStandardUser: {
			portaineree.OperationK8sResourcePoolsR:                   true,
			portaineree.OperationK8sResourcePoolDetailsR:             true,
			portaineree.OperationK8sApplicationsR:                    true,
			portaineree.OperationK8sApplicationsW:                    true,
			portaineree.OperationK8sApplicationDetailsR:              true,
			portaineree.OperationK8sApplicationDetailsW:              true,
			portaineree.OperationK8sApplicationsAdvancedDeploymentRW: true,
			portaineree.OperationK8sPodDelete:                        true,
			portaineree.OperationK8sApplicationConsoleRW:             true,
			portaineree.OperationK8sConfigurationsR:                  true,
			portaineree.OperationK8sConfigurationsW:                  true,
			portaineree.OperationK8sConfigurationsDetailsR:           true,
			portaineree.OperationK8sConfigurationsDetailsW:           true,
			portaineree.OperationK8sVolumesR:                         true,
			portaineree.OperationK8sVolumesW:                         true,
			portaineree.OperationK8sVolumeDetailsR:                   true,
			portaineree.OperationK8sVolumeDetailsW:                   true,
		},
		portaineree.RoleIDReadonly: {
			portaineree.OperationK8sResourcePoolsR:         true,
			portaineree.OperationK8sResourcePoolDetailsR:   true,
			portaineree.OperationK8sApplicationsR:          true,
			portaineree.OperationK8sApplicationDetailsR:    true,
			portaineree.OperationK8sConfigurationsR:        true,
			portaineree.OperationK8sConfigurationsDetailsR: true,
			portaineree.OperationK8sVolumesR:               true,
			portaineree.OperationK8sVolumeDetailsR:         true,
		},
	}
}

// DefaultK8sNamespaceAuthorizations returns a set of default k8s namespace-level authorizations
// based on user's role. The operations are supposed to be used in front-end.
func DefaultK8sNamespaceAuthorizations() map[portaineree.RoleID]portaineree.Authorizations {
	return map[portaineree.RoleID]portaineree.Authorizations{
		portaineree.RoleIDEndpointAdmin: {
			portaineree.OperationK8sAccessNamespaceRead:  true,
			portaineree.OperationK8sAccessNamespaceWrite: true,
		},
		portaineree.RoleIDHelpdesk: {
			portaineree.OperationK8sAccessNamespaceRead: true,
		},
		portaineree.RoleIDOperator: {
			portaineree.OperationK8sAccessNamespaceRead:  true,
			portaineree.OperationK8sAccessNamespaceWrite: true,
		},
		portaineree.RoleIDStandardUser: {
			portaineree.OperationK8sAccessNamespaceRead:  true,
			portaineree.OperationK8sAccessNamespaceWrite: true,
		},
		portaineree.RoleIDReadonly: {
			portaineree.OperationK8sAccessNamespaceRead: true,
		},
	}
}
