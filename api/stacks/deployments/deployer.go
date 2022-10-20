package deployments

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer/api/filesystem"
)

var (
	DefaultUnpackerImage = "portainer/compose-unpacker:latest"
	composePathPrefix    = "portainer-compose-unpacker"
)

type StackDeployer interface {
	DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error
	DeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error
	DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error
	DeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error
	UndeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
	UndeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
	DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error
	StartRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
	StopRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
	StartRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
	StopRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
}

type stackDeployer struct {
	lock                *sync.Mutex
	swarmStackManager   portaineree.SwarmStackManager
	composeStackManager portaineree.ComposeStackManager
	kubernetesDeployer  portaineree.KubernetesDeployer
	ClientFactory       *docker.ClientFactory
}

// NewStackDeployer inits a stackDeployer struct with a SwarmStackManager, a ComposeStackManager and a KubernetesDeployer
func NewStackDeployer(swarmStackManager portaineree.SwarmStackManager, composeStackManager portaineree.ComposeStackManager,
	kubernetesDeployer portaineree.KubernetesDeployer, clientFactory *docker.ClientFactory) *stackDeployer {
	return &stackDeployer{
		lock:                &sync.Mutex{},
		swarmStackManager:   swarmStackManager,
		composeStackManager: composeStackManager,
		kubernetesDeployer:  kubernetesDeployer,
		ClientFactory:       clientFactory,
	}
}

func (d *stackDeployer) DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)

	return d.swarmStackManager.Deploy(stack, prune, pullImage, endpoint)
}

func (d *stackDeployer) DeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)

	args := make(map[string]interface{})
	args["operation"] = "swarm-deploy"
	args["pullImage"] = pullImage
	args["prune"] = prune
	return d.remoteStack(stack, endpoint, args)
}

func (d *stackDeployer) UndeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	args := make(map[string]interface{})
	args["operation"] = "swarm-undeploy"
	return d.remoteStack(stack, endpoint, args)
}

func (d *stackDeployer) StartRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "swarm-start"
	return d.remoteStack(stack, endpoint, args)
}

func (d *stackDeployer) StopRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "swarm-stop"
	return d.remoteStack(stack, endpoint, args)
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
	err := d.composeStackManager.Up(context.TODO(), stack, endpoint, forceRereate)
	if err != nil {
		d.composeStackManager.Down(context.TODO(), stack, endpoint)
	}
	return err
}

func (d *stackDeployer) DeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error {
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
	args := make(map[string]interface{})
	args["operation"] = "deploy"
	return d.remoteStack(stack, endpoint, args)
}

func (d *stackDeployer) UndeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	args := make(map[string]interface{})
	args["operation"] = "undeploy"
	return d.remoteStack(stack, endpoint, args)
}

func (d *stackDeployer) StartRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "compose-start"
	return d.remoteStack(stack, endpoint, args)
}

func (d *stackDeployer) StopRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "compose-stop"
	return d.remoteStack(stack, endpoint, args)
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

	tokenData := &portaineree.TokenData{
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

func (d *stackDeployer) remoteStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, maps map[string]interface{}) error {
	cli, err := d.ClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		return errors.Wrap(err, "unable to create Docker client")
	}

	ctx := context.TODO()
	reader, err := cli.ImagePull(ctx, DefaultUnpackerImage, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to pull unpacker image")
	}

	defer reader.Close()
	io.Copy(ioutil.Discard, reader)

	info, err := cli.Info(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get agent info")
	}

	composeDestination := stack.ProjectPath
	if stack.FilesystemPath != "" {
		composeDestination = filesystem.JoinPaths(stack.FilesystemPath, composePathPrefix)
	} else {
		composeDestination = filesystem.JoinPaths(stack.ProjectPath, composePathPrefix)
	}

	targetSocketBind := "//./pipe/docker_engine"
	if strings.EqualFold(info.OSType, "linux") {
		targetSocketBind = "/var/run/docker.sock"
	}

	stackOperation := maps["operation"].(string)
	cmd := []string{}
	switch stackOperation {
	case "deploy", "undeploy":
		cmd = append(cmd, stackOperation)
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Username) != 0 && len(stack.GitConfig.Authentication.Password) != 0 {
			cmd = append(cmd, "-u")
			cmd = append(cmd, stack.GitConfig.Authentication.Username)
			cmd = append(cmd, "-p")
			cmd = append(cmd, stack.GitConfig.Authentication.Password)
		}
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
	case "compose-start":
		cmd = append(cmd, "deploy")
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Username) != 0 && len(stack.GitConfig.Authentication.Password) != 0 {
			cmd = append(cmd, "-u")
			cmd = append(cmd, stack.GitConfig.Authentication.Username)
			cmd = append(cmd, "-p")
			cmd = append(cmd, stack.GitConfig.Authentication.Password)
		}
		cmd = append(cmd, "-k")
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
	case "compose-stop":
		cmd = append(cmd, "undeploy")
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Username) != 0 && len(stack.GitConfig.Authentication.Password) != 0 {
			cmd = append(cmd, "-u")
			cmd = append(cmd, stack.GitConfig.Authentication.Username)
			cmd = append(cmd, "-p")
			cmd = append(cmd, stack.GitConfig.Authentication.Password)
		}
		cmd = append(cmd, "-k")
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
	case "swarm-deploy":
		cmd = append(cmd, stackOperation)
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Username) != 0 && len(stack.GitConfig.Authentication.Password) != 0 {
			cmd = append(cmd, "-u")
			cmd = append(cmd, stack.GitConfig.Authentication.Username)
			cmd = append(cmd, "-p")
			cmd = append(cmd, stack.GitConfig.Authentication.Password)
		}
		pullImage := maps["pullImage"].(bool)
		if pullImage {
			cmd = append(cmd, "-f")
		}
		prune := maps["prune"].(bool)
		if prune {
			cmd = append(cmd, "-r")
		}
		if len(stack.Env) > 0 {
			env := "--env=\""
			for _, pair := range stack.Env {
				env += pair.Name + "=" + pair.Value + ";"
			}
			env = env[0 : len(env)-1]
			env += "\""
			cmd = append(cmd, env)
		}
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
	case "swarm-undeploy":
		cmd = append(cmd, stackOperation)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
	case "swarm-stop":
		cmd = append(cmd, "swarm-undeploy")
		cmd = append(cmd, "-k")
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
	case "swarm-start":
		cmd = append(cmd, "swarm-deploy")
		cmd = append(cmd, "-k")
		cmd = append(cmd, "-f")
		cmd = append(cmd, "-r")
		if len(stack.Env) > 0 {
			env := "--env=\""
			for _, pair := range stack.Env {
				env += pair.Name + "=" + pair.Value + ";"
			}
			env = env[0 : len(env)-1]
			env += "\""
			cmd = append(cmd, env)
		}
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
	}

	for i := 0; i < len(stack.AdditionalFiles); i++ {
		cmd = append(cmd, stack.AdditionalFiles[i])
	}

	rand.Seed(time.Now().UnixNano())
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: DefaultUnpackerImage,
		Cmd:   cmd,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:%s", composeDestination, composeDestination),
			fmt.Sprintf("%s:%s", targetSocketBind, targetSocketBind),
		},
	}, nil, nil, fmt.Sprintf("portainer-unpacker-%d-%s-%d", stack.ID, stack.Name, rand.Intn(100)))
	defer cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to create unpacker container")
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrap(err, "start unpacker container error")
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return errors.Wrap(err, "An error occured while waiting for the deployment of the stack.")
		}
	case <-statusCh:
	}
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err == nil {
		outputBytes, err := ioutil.ReadAll(out)
		if err == nil {
			log.Printf(string(outputBytes))
		}
	}

	status, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return errors.Wrap(err, "fetch container information error")
	}

	if status.State.ExitCode != 0 {
		return fmt.Errorf("An error occured while running unpacker container with exit code %d", status.State.ExitCode)
	}

	return nil
}
