package deployments

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/docker/client"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer/api/filesystem"
)

var (
	defaultUnpackerImage       = "portainer/compose-unpacker:latest"
	composeUnpackerImageEnvVar = "COMPOSE_UNPACKER_IMAGE"
	composePathPrefix          = "portainer-compose-unpacker"
)

type StackDeployer interface {
	DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error
	DeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error
	DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error
	DeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRereate bool) error
	UndeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
	UndeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error
	DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error
	RestartKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User, resourceList []string) error
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

func (d *stackDeployer) DeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)

	args := make(map[string]interface{})
	args["operation"] = "swarm-deploy"
	args["pullImage"] = pullImage
	args["prune"] = prune
	return d.remoteStack(stack, endpoint, args, registries)
}

func (d *stackDeployer) UndeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	args := make(map[string]interface{})
	args["operation"] = "swarm-undeploy"
	return d.remoteStack(stack, endpoint, args, nil)
}

func (d *stackDeployer) StartRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "swarm-start"
	return d.remoteStack(stack, endpoint, args, nil)
}

func (d *stackDeployer) StopRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "swarm-stop"
	return d.remoteStack(stack, endpoint, args, nil)
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
	return d.remoteStack(stack, endpoint, args, registries)
}

func (d *stackDeployer) UndeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	args := make(map[string]interface{})
	args["operation"] = "undeploy"
	return d.remoteStack(stack, endpoint, args, nil)
}

func (d *stackDeployer) StartRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "compose-start"
	return d.remoteStack(stack, endpoint, args, nil)
}

func (d *stackDeployer) StopRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	args := make(map[string]interface{})
	args["operation"] = "compose-stop"
	return d.remoteStack(stack, endpoint, args, nil)
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

	tokenData := &portaineree.TokenData{
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

// remoteStack is used to deploy a stack on a remote endpoint based on the supplied `maps` of arguments
//
// it deploys a container of https://github.com/portainer/compose-unpacker on the remote endpoint with a set of command arguments
func (d *stackDeployer) remoteStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, maps map[string]interface{}, registries []portaineree.Registry) error {
	ctx := context.TODO()

	cli, err := d.createDockerClient(ctx, endpoint)
	if err != nil {
		return errors.WithMessage(err, "unable to create docker client")
	}
	defer cli.Close()

	image := getUnpackerImage()

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to pull unpacker image")
	}

	defer reader.Close()
	io.Copy(ioutil.Discard, reader)

	info, err := cli.Info(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get agent info")
	}

	var composeDestination string
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
	// deploy [-u username -p password] [-k] [--env KEY1=VALUE1 --env KEY2=VALUE2] <git-repo-url> <ref> <project-name> <destination> <compose-file-path> [<more-file-paths>...]
	case "deploy":
		cmd = append(cmd, stackOperation)
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Password) != 0 {
			cmd = append(cmd, "-u")
			cmd = append(cmd, stack.GitConfig.Authentication.Username)
			cmd = append(cmd, "-p")
			cmd = append(cmd, stack.GitConfig.Authentication.Password)
		}
		cmd = append(cmd, getEnv(stack.Env)...)
		cmd = append(cmd, getRegistry(registries, d.dataStore)...)
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.GitConfig.ReferenceName)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
		for i := 0; i < len(stack.AdditionalFiles); i++ {
			cmd = append(cmd, stack.AdditionalFiles[i])
		}
	// undeploy [-u username -p password] [-k] <git-repo-url> <project-name> <destination> <compose-file-path> [<more-file-paths>...]
	case "undeploy":
		cmd = append(cmd, stackOperation)
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Password) != 0 {
			cmd = append(cmd, "-u")
			cmd = append(cmd, stack.GitConfig.Authentication.Username)
			cmd = append(cmd, "-p")
			cmd = append(cmd, stack.GitConfig.Authentication.Password)
		}
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
		for i := 0; i < len(stack.AdditionalFiles); i++ {
			cmd = append(cmd, stack.AdditionalFiles[i])
		}
	case "compose-start":
		// deploy [-u username -p password] [-k] [--env KEY1=VALUE1 --env KEY2=VALUE2] <git-repo-url> <project-name> <destination> <compose-file-path> [<more-file-paths>...]
		cmd = append(cmd, "deploy")
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Password) != 0 {
			cmd = append(cmd, "-u")
			cmd = append(cmd, stack.GitConfig.Authentication.Username)
			cmd = append(cmd, "-p")
			cmd = append(cmd, stack.GitConfig.Authentication.Password)
		}
		cmd = append(cmd, "-k")
		cmd = append(cmd, getEnv(stack.Env)...)
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
		for i := 0; i < len(stack.AdditionalFiles); i++ {
			cmd = append(cmd, stack.AdditionalFiles[i])
		}
	case "compose-stop":
		// undeploy [-u username -p password] [-k] <git-repo-url> <project-name> <destination> <compose-file-path> [<more-file-paths>...]
		cmd = append(cmd, "undeploy")
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Password) != 0 {
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
		for i := 0; i < len(stack.AdditionalFiles); i++ {
			cmd = append(cmd, stack.AdditionalFiles[i])
		}
	case "swarm-deploy":
		// deploy [-u username -p password] [-f] [-r] [-k] [--env KEY1=VALUE1 --env KEY2=VALUE2] <git-repo-url> <git-ref> <project-name> <destination> <compose-file-path> [<more-file-paths>...]
		cmd = append(cmd, stackOperation)
		if stack.GitConfig.Authentication != nil && len(stack.GitConfig.Authentication.Password) != 0 {
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
		cmd = append(cmd, getEnv(stack.Env)...)
		cmd = append(cmd, getRegistry(registries, d.dataStore)...)
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.GitConfig.ReferenceName)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
		for i := 0; i < len(stack.AdditionalFiles); i++ {
			cmd = append(cmd, stack.AdditionalFiles[i])
		}
	case "swarm-undeploy":
		// undeploy [-k] <project-name> <destination>
		cmd = append(cmd, stackOperation)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
	case "swarm-stop":
		// undeploy [-k] <project-name> <destination>
		cmd = append(cmd, "swarm-undeploy")
		cmd = append(cmd, "-k")
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
	case "swarm-start":
		// deploy [-u username -p password] [-f] [-r] [-k] [--env KEY1=VALUE1 --env KEY2=VALUE2] <git-repo-url> <project-name> <destination> <compose-file-path> [<more-file-paths>...]
		cmd = append(cmd, "swarm-deploy")
		cmd = append(cmd, "-k")
		cmd = append(cmd, "-f")
		cmd = append(cmd, "-r")
		cmd = append(cmd, getEnv(stack.Env)...)
		cmd = append(cmd, stack.GitConfig.URL)
		cmd = append(cmd, stack.Name)
		cmd = append(cmd, composeDestination)
		cmd = append(cmd, stack.EntryPoint)
		for i := 0; i < len(stack.AdditionalFiles); i++ {
			cmd = append(cmd, stack.AdditionalFiles[i])
		}
	}

	log.Debug().
		Str("image", image).
		Str("cmd", strings.Join(cmd, " ")).
		Msg("running unpacker")

	rand.Seed(time.Now().UnixNano())
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
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
			return errors.Wrap(err, "An error occurred while waiting for the deployment of the stack.")
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Error().Err(err).Msg("unable to get logs from unpacker container")
	} else {
		outputBytes, err := ioutil.ReadAll(out)
		if err != nil {
			log.Error().Err(err).Msg("unable to parse logs from unpacker container")
		} else {
			log.Info().
				Str("output", string(outputBytes)).
				Msg("Stack deployment output")
		}
	}

	status, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return errors.Wrap(err, "fetch container information error")
	}

	if status.State.ExitCode != 0 {
		return fmt.Errorf("An error occurred while running unpacker container with exit code %d", status.State.ExitCode)
	}

	return nil
}

func (d *stackDeployer) createDockerClient(ctx context.Context, endpoint *portaineree.Endpoint) (*dockerclient.Client, error) {
	cli, err := d.ClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Docker client")
	}

	// only for swarm
	info, err := cli.Info(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get agent info")
	}

	// if swarm - create client for swarm leader
	if info.Swarm.LocalNodeState == swarm.LocalNodeStateInactive {
		return cli, nil
	}
	nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list nodes")
	}

	if len(nodes) == 0 {
		return nil, errors.New("no nodes available")
	}

	var managerNode swarm.Node
	for _, node := range nodes {
		if node.ManagerStatus != nil && node.ManagerStatus.Leader {
			managerNode = node
			break
		}
	}

	if managerNode.ID == "" {
		return nil, errors.New("no leader node available")
	}

	cli.Close()
	return d.ClientFactory.CreateClient(endpoint, managerNode.Description.Hostname, nil)
}

func getEnv(env []portaineree.Pair) []string {
	if len(env) == 0 {
		return nil
	}

	cmd := []string{}
	for _, pair := range env {
		cmd = append(cmd, fmt.Sprintf(`--env=%s=%s`, pair.Name, pair.Value))
	}

	return cmd
}

func getRegistry(registries []portaineree.Registry, dataStore dataservices.DataStore) []string {
	cmds := []string{}

	for _, registry := range registries {
		if registry.Authentication {
			err := registryutils.EnsureRegTokenValid(dataStore, &registry)
			if err == nil {
				username, password, err := registryutils.GetRegEffectiveCredential(&registry)
				if err == nil {
					cmd := fmt.Sprintf("--registry=%s:%s:%s", username, password, registry.URL)
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return cmds
}

func getUnpackerImage() string {
	image := os.Getenv(composeUnpackerImageEnvVar)
	if image == "" {
		image = defaultUnpackerImage
	}

	return image
}
