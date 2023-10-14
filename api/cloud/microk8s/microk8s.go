package microk8s

import (
	"bytes"
	"fmt"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

func GetCurrentVersion(datastore dataservices.DataStore, credential *models.CloudCredential, environmentID portainer.EndpointID) (string, error) {
	log.Debug().Str(
		"provider",
		portaineree.CloudProviderMicrok8s,
	).Msg("processing version request")

	// Gather nodeIP from environmentID.
	endpoint, err := datastore.Endpoint().Endpoint(portainer.EndpointID(environmentID))
	if err != nil {
		log.Debug().Str(
			"provider",
			portaineree.CloudProviderMicrok8s,
		).Msg("failed looking up environment nodeIP")
		return "", err
	}
	nodeIP, _, _ := strings.Cut(endpoint.URL, ":")

	// Gather current version.
	// We need to ssh into the server to fetch this live. Even if we stored the
	// version in the database, it could be outdated as the user can always
	// update their cluster manually outside of portainer.
	sshClient, err := sshUtil.NewConnectionWithCredentials(nodeIP, credential)
	if err != nil {
		log.Debug().Err(err).Msg("failed creating ssh credentials")
		return "", err
	}
	defer sshClient.Close()

	// We can't use the microk8s version command as it was added in 1.25.
	// Instead we parse the output from snap.
	var resp bytes.Buffer
	err = sshClient.RunCommand(
		"snap list",
		&resp,
	)
	if err != nil {
		return "", fmt.Errorf("error running ssh command: %w", err)
	}
	return ParseSnapInstalledVersion(resp.String())
}
