package cloud

import (
	"github.com/gofrs/uuid"
	portainer "github.com/portainer/portainer/api"
)

type PreinstalledAgentProvisioningClusterRequest struct {
	EnvironmentID portainer.EndpointID `json:"environmentID"`
}

func (service *CloudManagementService) PreinstalledAgentProvisionCluster(req PreinstalledAgentProvisioningClusterRequest) (string, error) {
	// The agent is already installed so we don't really need to do anything here.
	uid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

func (service *CloudManagementService) PreinstalledAgentGetCluster(id string) (*KaasCluster, error) {
	return &KaasCluster{
		Id:    id,
		Ready: true,
	}, nil
}
