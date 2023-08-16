package cloud

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/rs/zerolog/log"

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

		// Copy what's needed because endpoint is deleted while this is in progress
		url := microk8s.UrlToMasterNode(endpoint.URL)
		name := endpoint.Name
		endpoint = nil // Avoid using endpoint by mistake now

		// background delete microk8s because it's quite slow!
		go func() {
			err := microk8s.DeleteCluster(url, credentials)
			if err != nil {
				log.Warn().Msgf("Failed to clean-up some microk8s nodes when deleting the cluster. %v", err)
			}

			log.Info().Msgf("Successfully removed microk8s from cluster %s", name)
		}()

		return nil
	}

	return fmt.Errorf("delete cluster not supported for this provider: %s", endpoint.CloudProvider.Provider)
}
