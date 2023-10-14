package authorization

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

// DefaultK8sClusterAuthorizations returns a set of default k8s cluster-level authorizations
// based on user's role. The operations are supposed to be used in front-end.
func DefaultK8sClusterAuthorizations() map[portainer.RoleID]portainer.Authorizations {
	return map[portainer.RoleID]portainer.Authorizations{
		portaineree.RoleIDEndpointAdmin: {
			portaineree.OperationK8sAccessAllNamespaces:              true,
			portaineree.OperationK8sAccessSystemNamespaces:           true,
			portaineree.OperationK8sAccessUserNamespaces:             true,
			portaineree.OperationK8sResourcePoolsR:                   true,
			portaineree.OperationK8sResourcePoolsW:                   true,
			portaineree.OperationK8sResourcePoolDetailsR:             true,
			portaineree.OperationK8sResourcePoolDetailsW:             true,
			portaineree.OperationK8sResourcePoolsAccessManagementRW:  true,
			portaineree.OperationK8sPodSecurityW:                     true,
			portaineree.OperationK8sApplicationsR:                    true,
			portaineree.OperationK8sApplicationsW:                    true,
			portaineree.OperationK8sApplicationDetailsR:              true,
			portaineree.OperationK8sApplicationDetailsW:              true,
			portaineree.OperationK8sApplicationP:                     true,
			portaineree.OperationK8sPodDelete:                        true,
			portaineree.OperationK8sApplicationConsoleRW:             true,
			portaineree.OperationK8sApplicationsAdvancedDeploymentRW: true,
			portaineree.OperationK8sConfigMapsR:                      true,
			portaineree.OperationK8sConfigMapsW:                      true,
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
			portaineree.OperationK8sIngressesR:                       true,
			portaineree.OperationK8sIngressesW:                       true,
			portaineree.OperationK8sYAMLW:                            true,
			portaineree.OperationK8sSecretsR:                         true,
			portaineree.OperationK8sSecretsW:                         true,
			portaineree.OperationK8sServicesR:                        true,
			portaineree.OperationK8sServicesW:                        true,
			portaineree.OperationK8sServiceAccountsR:                 true,
			portaineree.OperationK8sServiceAccountsW:                 true,
			portaineree.OperationK8sClusterRolesR:                    true,
			portaineree.OperationK8sClusterRolesW:                    true,
			portaineree.OperationK8sClusterRoleBindingsR:             true,
			portaineree.OperationK8sClusterRoleBindingsW:             true,
			portaineree.OperationK8sRolesR:                           true,
			portaineree.OperationK8sRolesW:                           true,
			portaineree.OperationK8sRoleBindingsR:                    true,
			portaineree.OperationK8sRoleBindingsW:                    true,
		},
		portaineree.RoleIDOperator: {
			portaineree.OperationK8sAccessUserNamespaces:     true,
			portaineree.OperationK8sResourcePoolsR:           true,
			portaineree.OperationK8sResourcePoolDetailsR:     true,
			portaineree.OperationK8sApplicationsR:            true,
			portaineree.OperationK8sApplicationDetailsR:      true,
			portaineree.OperationK8sApplicationP:             true,
			portaineree.OperationK8sPodDelete:                true,
			portaineree.OperationK8sApplicationConsoleRW:     true,
			portaineree.OperationK8sConfigMapsR:              true,
			portaineree.OperationK8sConfigMapsW:              true,
			portaineree.OperationK8sVolumesR:                 true,
			portaineree.OperationK8sVolumeDetailsR:           true,
			portaineree.OperationK8sClusterR:                 true,
			portaineree.OperationK8sClusterNodeR:             true,
			portaineree.OperationK8sApplicationErrorDetailsR: true,
			portaineree.OperationK8sStorageClassDisabledR:    true,
			portaineree.OperationK8sIngressesR:               true,
			portaineree.OperationK8sYAMLR:                    true,
			portaineree.OperationK8sSecretsR:                 true,
			portaineree.OperationK8sServicesR:                true,
		},
		portaineree.RoleIDHelpdesk: {
			portaineree.OperationK8sAccessUserNamespaces:     true,
			portaineree.OperationK8sResourcePoolsR:           true,
			portaineree.OperationK8sResourcePoolDetailsR:     true,
			portaineree.OperationK8sApplicationsR:            true,
			portaineree.OperationK8sApplicationDetailsR:      true,
			portaineree.OperationK8sConfigMapsR:              true,
			portaineree.OperationK8sVolumesR:                 true,
			portaineree.OperationK8sVolumeDetailsR:           true,
			portaineree.OperationK8sClusterR:                 true,
			portaineree.OperationK8sClusterNodeR:             true,
			portaineree.OperationK8sApplicationErrorDetailsR: true,
			portaineree.OperationK8sStorageClassDisabledR:    true,
			portaineree.OperationK8sIngressesR:               true,
			portaineree.OperationK8sYAMLR:                    true,
			portaineree.OperationK8sSecretsR:                 true,
			portaineree.OperationK8sServicesR:                true,
		},
		portaineree.RoleIDStandardUser: {
			portaineree.OperationK8sResourcePoolsR:                   true,
			portaineree.OperationK8sResourcePoolDetailsR:             true,
			portaineree.OperationK8sApplicationsR:                    true,
			portaineree.OperationK8sApplicationsW:                    true,
			portaineree.OperationK8sApplicationDetailsR:              true,
			portaineree.OperationK8sApplicationDetailsW:              true,
			portaineree.OperationK8sApplicationP:                     true,
			portaineree.OperationK8sApplicationsAdvancedDeploymentRW: true,
			portaineree.OperationK8sPodDelete:                        true,
			portaineree.OperationK8sApplicationConsoleRW:             true,
			portaineree.OperationK8sConfigMapsR:                      true,
			portaineree.OperationK8sConfigMapsW:                      true,
			portaineree.OperationK8sVolumesR:                         true,
			portaineree.OperationK8sVolumesW:                         true,
			portaineree.OperationK8sVolumeDetailsR:                   true,
			portaineree.OperationK8sVolumeDetailsW:                   true,
			portaineree.OperationK8sIngressesR:                       true,
			portaineree.OperationK8sIngressesW:                       true,
			portaineree.OperationK8sYAMLW:                            true,
			portaineree.OperationK8sSecretsR:                         true,
			portaineree.OperationK8sSecretsW:                         true,
			portaineree.OperationK8sServicesR:                        true,
			portaineree.OperationK8sServicesW:                        true,
		},
		portaineree.RoleIDReadonly: {
			portaineree.OperationK8sResourcePoolsR:       true,
			portaineree.OperationK8sResourcePoolDetailsR: true,
			portaineree.OperationK8sApplicationsR:        true,
			portaineree.OperationK8sApplicationDetailsR:  true,
			portaineree.OperationK8sConfigMapsR:          true,
			portaineree.OperationK8sVolumesR:             true,
			portaineree.OperationK8sVolumeDetailsR:       true,
			portaineree.OperationK8sIngressesR:           true,
			portaineree.OperationK8sYAMLR:                true,
			portaineree.OperationK8sSecretsR:             true,
			portaineree.OperationK8sClusterRolesR:        true,
		},
	}
}

// DefaultK8sNamespaceAuthorizations returns a set of default k8s namespace-level authorizations
// based on user's role. The operations are supposed to be used in front-end.
func DefaultK8sNamespaceAuthorizations() map[portainer.RoleID]portainer.Authorizations {
	return map[portainer.RoleID]portainer.Authorizations{
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
