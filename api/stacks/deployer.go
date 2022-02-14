package stacks

import (
	"context"
	"os"
	"sync"

	"github.com/pkg/errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/stackutils"
	k "github.com/portainer/portainer-ee/api/kubernetes"
)

type StackDeployer interface {
	DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error
	DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error
	DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error
}

type stackDeployer struct {
	lock                *sync.Mutex
	swarmStackManager   portaineree.SwarmStackManager
	composeStackManager portaineree.ComposeStackManager
	kubernetesDeployer  portaineree.KubernetesDeployer
}

// NewStackDeployer inits a stackDeployer struct with a SwarmStackManager, a ComposeStackManager and a KubernetesDeployer
func NewStackDeployer(swarmStackManager portaineree.SwarmStackManager, composeStackManager portaineree.ComposeStackManager, kubernetesDeployer portaineree.KubernetesDeployer) *stackDeployer {
	return &stackDeployer{
		lock:                &sync.Mutex{},
		swarmStackManager:   swarmStackManager,
		composeStackManager: composeStackManager,
		kubernetesDeployer:  kubernetesDeployer,
	}
}

func (d *stackDeployer) DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)

	return d.swarmStackManager.Deploy(stack, prune, pullImage, endpoint)
}

func (d *stackDeployer) DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error {
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
	return d.composeStackManager.Up(context.TODO(), stack, endpoint, forceRereate)
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

	manifestFilePaths, tempDir, err := stackutils.CreateTempK8SDeploymentFiles(stack, d.kubernetesDeployer, appLabels)
	if err != nil {
		return errors.Wrap(err, "failed to create temp kub deployment files")
	}
	defer os.RemoveAll(tempDir)

	_, err = d.kubernetesDeployer.Deploy(user.ID, endpoint, manifestFilePaths, stack.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to deploy kubernetes application")
	}

	return nil
}
