package exectest

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

type kubernetesMockDeployer struct{}

// NewKubernetesDeployer creates a mock kubernetes deployer
func NewKubernetesDeployer() portaineree.KubernetesDeployer {
	return &kubernetesMockDeployer{}
}

func (deployer *kubernetesMockDeployer) Deploy(userID portainer.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return "", nil
}

func (deployer *kubernetesMockDeployer) Restart(userID portainer.UserID, endpoint *portaineree.Endpoint, resourceList []string, namespace string) (string, error) {
	return "", nil
}

func (deployer *kubernetesMockDeployer) DeployViaKubeConfig(kubeConfig string, clusterID string, manifestFile string) error {
	return nil
}

func (deployer *kubernetesMockDeployer) Remove(userID portainer.UserID, endpoint *portaineree.Endpoint, manifestFiles []string, namespace string) (string, error) {
	return "", nil
}

func (deployer *kubernetesMockDeployer) ConvertCompose(data []byte) ([]byte, error) {
	return nil, nil
}
