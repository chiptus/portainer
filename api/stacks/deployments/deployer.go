package deployments

import (
	"context"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker/client"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
)

type BaseStackDeployer interface {
	DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error
	DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRecreate bool) error
	DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error
	RestartKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User, resourceList []string) error
}

type StackDeployer interface {
	BaseStackDeployer
	RemoteStackDeployer
}

type stackDeployer struct {
	lock                *sync.Mutex
	swarmStackManager   portaineree.SwarmStackManager
	composeStackManager portaineree.ComposeStackManager
	kubernetesDeployer  portaineree.KubernetesDeployer
	ClientFactory       *client.ClientFactory
	dataStore           dataservices.DataStore
}

// NewStackDeployer inits a stackDeployer struct with a SwarmStackManager, a ComposeStackManager and a KubernetesDeployer
func NewStackDeployer(swarmStackManager portaineree.SwarmStackManager, composeStackManager portaineree.ComposeStackManager,
	kubernetesDeployer portaineree.KubernetesDeployer, clientFactory *client.ClientFactory, dataStore dataservices.DataStore) *stackDeployer {
	return &stackDeployer{
		lock:                &sync.Mutex{},
		swarmStackManager:   swarmStackManager,
		composeStackManager: composeStackManager,
		kubernetesDeployer:  kubernetesDeployer,
		ClientFactory:       clientFactory,
		dataStore:           dataStore,
	}
}

func (d *stackDeployer) DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)

	return d.swarmStackManager.Deploy(stack, prune, pullImage, endpoint)
}

func (d *stackDeployer) DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRecreate bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)
	// --force-recreate doesn't pull updated images
	if forcePullImage {
		err := d.composeStackManager.Pull(context.TODO(), stack, endpoint)
		if err != nil {
			return err
		}
	}

	err := d.composeStackManager.Up(context.TODO(), stack, endpoint, forceRecreate)
	if err != nil {
		d.composeStackManager.Down(context.TODO(), stack, endpoint)
	}
	return err
}

func (d *stackDeployer) DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	appLabels := k.KubeAppLabels{
		StackID:   int(stack.ID),
		StackName: stack.Name,
		Owner:     user.Username,
	}

	if stack.GitConfig == nil {
		appLabels.Kind = "content"
	} else {
		appLabels.Kind = "git"
	}

	tokenData := &portainer.TokenData{
		ID: user.ID,
	}

	k8sDeploymentConfig, err := CreateKubernetesStackDeploymentConfig(stack, d.kubernetesDeployer, appLabels, tokenData, endpoint, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create temp kub deployment files")
	}

	err = k8sDeploymentConfig.Deploy()
	if err != nil {
		return errors.Wrap(err, "failed to deploy kubernetes application")
	}

	return nil
}

// Restart Kubernetes Stack.  If the resource list is empty, use all the resources from the stack yaml files.
// If the provided resources don't exist, don't pass them down to the cli
func (d *stackDeployer) RestartKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User, resourceList []string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	appLabels := k.KubeAppLabels{
		StackID:   int(stack.ID),
		StackName: stack.Name,
		Owner:     user.Username,
	}

	if stack.GitConfig == nil {
		appLabels.Kind = "content"
	} else {
		appLabels.Kind = "git"
	}

	tokenData := &portainer.TokenData{
		ID: user.ID,
	}

	k8sDeploymentConfig, err := CreateKubernetesStackDeploymentConfig(stack, d.kubernetesDeployer, appLabels, tokenData, endpoint, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create temp kub deployment files")
	}

	err = k8sDeploymentConfig.Restart(resourceList)
	if err != nil {
		return errors.Wrap(err, "failed to deploy kubernetes application")
	}

	return nil
}
