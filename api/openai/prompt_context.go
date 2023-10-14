package openai

import (
	"fmt"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

// buildServerAwareContext returns a context that is aware of the Portainer EE server.
func buildServerAwareContext() string {
	return fmt.Sprintf("I am using Portainer EE version %s.", portaineree.APIVersion)
}

// buildEnvironmentAwareContext returns a context that is aware of the specified environment.
func (builder *OpenAIPromptBuilder) buildEnvironmentAwareContext(environmentID portainer.EndpointID) (string, error) {
	ctx := []string{}

	environment, err := builder.DataStore.Endpoint().Endpoint(environmentID)
	if err != nil {
		log.Error().Err(err).Msg("unable to retrieve environment from the database")
		return "", err
	}

	err = builder.SnapshotService.FillSnapshotData(environment)
	if err != nil {
		log.Error().Err(err).Msg("unable to inject snapshot data for the specified environment")
		return "", err
	}

	switch environment.Type {
	case portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment, portaineree.EdgeAgentOnDockerEnvironment:
		ctx = append(ctx, buildDockerAwareContext(environment))
	case portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment:
		ctx = append(ctx, buildKubernetesAwareContext(environment))
	}

	return strings.Join(ctx, " "), nil
}

// buildUserAwareContext returns a context that is aware of the specified user.
func (builder *OpenAIPromptBuilder) buildUserAwareContext(user portaineree.User) string {
	ctx := ""

	if user.Role == portaineree.AdministratorRole {
		ctx = "I am a Portainer Administrator."
	} else {
		ctx = "I am not a Portainer administrator."

		// We could potentially lookup for the user permission against the environment here
		// and add more context to the query. For example, "I am assigned the HelpDesk user role."
	}

	return ctx
}

func buildDockerAwareContext(environment *portaineree.Endpoint) string {
	ctx := []string{`Write the response as a user guide using Markdown format. Use the following format when I request help with application deployment, otherwise use the format of your choice for the answer:
## Introduction
Give a short description about how to solve my problem in this section.
## Docker Compose file
Optional section, write a Docker compose file if possible.
## Configuration
Optional section, give a very thorough explanation of the configuration available in the Docker compose file if one was specified.
## Portainer EE instructions
Write the instructions that I need to follow in Portainer EE.

Here is more information about my environment:`}
	if len(environment.Snapshots) > 0 {
		snapshot := environment.Snapshots[0]

		if snapshot.Swarm {
			ctx = append(ctx, fmt.Sprintf("I manage a Docker Swarm cluster of %d nodes.", snapshot.NodeCount))
		} else {
			ctx = append(ctx, fmt.Sprintf("I manage a single standalone Docker host using Docker version %s.", snapshot.DockerVersion))
		}
	}

	// We could potentially lookup for environment security settings here that are specific to Docker.
	// For example, "I cannot run containers with the --privileged flag."
	// We could even include live queries to the Docker environment.

	return strings.Join(ctx, " ")
}

func buildKubernetesAwareContext(environment *portaineree.Endpoint) string {
	ctx := []string{`Write the response as a user guide using Markdown format. Use the following format when I request help with application deployment, otherwise use the format of your choice for the answer:
## Introduction
Give a short description about how to solve my problem in this section.
## Manifest
Optional section, write a single Kubernetes manifest file containing all the required Kubernetes resources if possible.
## Configuration
Optional section, give a very thorough explanation of the configuration available in the Kubernetes manifest file if one was specified.
## Portainer EE instructions
Write the instructions that I need to follow in Portainer EE.

Here is more information about my environment:`}

	if len(environment.Kubernetes.Snapshots) > 0 {
		snapshot := environment.Kubernetes.Snapshots[0]
		ctx = append(ctx, fmt.Sprintf("I manage a Kubernetes cluster of %d nodes.", snapshot.NodeCount))
	}

	// We could potentially lookup for environment components here that are specific to Kubernetes.
	// For example, "I have a Kubernetes Ingress Controller installed."
	// We could even include live queries to the Kubernetes environment.

	return strings.Join(ctx, " ")
}
