package authorization

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// DefaultAzureAuthorizations returns a set of default azure authorizations based on user's role.
func DefaultAzureAuthorizations() map[portaineree.RoleID]portaineree.Authorizations {
	return map[portaineree.RoleID]portaineree.Authorizations{
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
