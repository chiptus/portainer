package cloud

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/portainer/portainer-ee/api/cloud/microk8s"
	"github.com/portainer/portainer-ee/api/dataservices"
)

func (service *CloudManagementService) DeleteCluster(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint) error {
	credentials, err := tx.CloudCredential().Read(endpoint.CloudProvider.CredentialID)
	if err != nil {
		return err
	}

	switch endpoint.CloudProvider.Provider {
	case portaineree.CloudProviderMicrok8s:
		return microk8s.DeleteCluster(endpoint, credentials)
	}

	return fmt.Errorf("delete cluster not supported for this provider: %s", endpoint.CloudProvider.Provider)
}
