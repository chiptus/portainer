package testhelpers

import portaineree "github.com/portainer/portainer-ee/api"

type testStackDeployer struct {
}

func NewTestStackDeployer() *testStackDeployer {
	return &testStackDeployer{}
}

func (testStackDeployer) DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	return nil
}

func (testStackDeployer) DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRecreate bool) error {
	return nil
}

func (testStackDeployer) DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error {
	return nil
}

func (testStackDeployer) RestartKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User, resourceList []string) error {
	return nil
}

func (testStackDeployer) DeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRecreate bool) error {
	return nil
}

func (testStackDeployer) UndeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	return nil
}

func (testStackDeployer) StartRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry) error {
	return nil
}

func (testStackDeployer) StopRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	return nil
}

func (testStackDeployer) DeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	return nil
}

func (testStackDeployer) UndeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	return nil
}

func (testStackDeployer) StartRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry) error {
	return nil
}

func (testStackDeployer) StopRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	return nil
}
