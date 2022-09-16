package exectest

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

type kubernetesMockDeployer struct{}

// NewKubernetesDeployer creates a mock kubernetes deployer
func NewKubernetesDeployer() portaineree.KubernetesDeployer {
	return &kubernetesMockDeployer{}
}

func (deployer *kubernetesMockDeployer) Deploy(userID portaineree.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return "", nil
}

func (deployer *kubernetesMockDeployer) Remove(userID portaineree.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return "", nil
}

func (deployer *kubernetesMockDeployer) ConvertCompose(data []byte) ([]byte, error) {
	return nil, nil
}
