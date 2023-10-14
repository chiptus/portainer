package authorization

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

// DefaultAzureAuthorizations returns a set of default azure authorizations based on user's role.
func DefaultAzureAuthorizations() map[portainer.RoleID]portainer.Authorizations {
	return map[portainer.RoleID]portainer.Authorizations{
		portaineree.RoleIDEndpointAdmin: {
			portaineree.OperationAzureSubscriptionsList:    true,
			portaineree.OperationAzureSubscriptionGet:      true,
			portaineree.OperationAzureProviderGet:          true,
			portaineree.OperationAzureResourceGroupsList:   true,
			portaineree.OperationAzureResourceGroupGet:     true,
			portaineree.OperationAzureContainerGroupsList:  true,
			portaineree.OperationAzureContainerGroupGet:    true,
			portaineree.OperationAzureContainerGroupCreate: true,
			portaineree.OperationAzureContainerGroupDelete: true,
		},
		portaineree.RoleIDOperator: {
			portaineree.OperationAzureSubscriptionsList:   true,
			portaineree.OperationAzureSubscriptionGet:     true,
			portaineree.OperationAzureProviderGet:         true,
			portaineree.OperationAzureResourceGroupsList:  true,
			portaineree.OperationAzureResourceGroupGet:    true,
			portaineree.OperationAzureContainerGroupsList: true,
			portaineree.OperationAzureContainerGroupGet:   true,
		},
		portaineree.RoleIDHelpdesk: {
			portaineree.OperationAzureSubscriptionsList:   true,
			portaineree.OperationAzureSubscriptionGet:     true,
			portaineree.OperationAzureProviderGet:         true,
			portaineree.OperationAzureResourceGroupsList:  true,
			portaineree.OperationAzureResourceGroupGet:    true,
			portaineree.OperationAzureContainerGroupsList: true,
			portaineree.OperationAzureContainerGroupGet:   true,
		},
		portaineree.RoleIDStandardUser: {
			portaineree.OperationAzureSubscriptionsList:    true,
			portaineree.OperationAzureSubscriptionGet:      true,
			portaineree.OperationAzureProviderGet:          true,
			portaineree.OperationAzureResourceGroupsList:   true,
			portaineree.OperationAzureResourceGroupGet:     true,
			portaineree.OperationAzureContainerGroupsList:  true,
			portaineree.OperationAzureContainerGroupGet:    true,
			portaineree.OperationAzureContainerGroupCreate: true,
			portaineree.OperationAzureContainerGroupDelete: true,
		},
		portaineree.RoleIDReadonly: {
			portaineree.OperationAzureSubscriptionsList:   true,
			portaineree.OperationAzureSubscriptionGet:     true,
			portaineree.OperationAzureProviderGet:         true,
			portaineree.OperationAzureResourceGroupsList:  true,
			portaineree.OperationAzureResourceGroupGet:    true,
			portaineree.OperationAzureContainerGroupsList: true,
			portaineree.OperationAzureContainerGroupGet:   true,
		},
	}
}
