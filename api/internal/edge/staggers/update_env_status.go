package staggers

import (
	"slices"
	"time"

	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

type UpdateEndpointStatusFunc func(int, portainer.EdgeStackStatus) *portainer.EdgeStackStatus

func updateEnvironmentStatus(dataStore dataservices.DataStore, edgeStackID portainer.EdgeStackID, updateEnvStatusFunc UpdateEndpointStatusFunc) {
	log.Debug().Msg("[Stagger update environment status] Updating environment status")

	err := dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		stack, err := tx.EdgeStack().EdgeStack(edgeStackID)
		if err != nil {
			if dataservices.IsErrObjectNotFound(err) {
				log.Warn().Err(err).Msg("[Stagger endpoint status update] Unable to find a stack inside the database, skipping error")
				return nil
			}
			return err
		}

		targetVersion := stack.StackFileVersion
		log.Debug().Int("targetVersion", targetVersion).Msg("[Stagger update environment status] Updating environment status")
		for endpointID, endpointStatus := range stack.Status {
			updatedEndpointStatus := updateEnvStatusFunc(targetVersion, endpointStatus)
			if updatedEndpointStatus == nil {
				continue
			}

			stack.Status[endpointID] = *updatedEndpointStatus
		}

		return tx.EdgeStack().UpdateEdgeStack(edgeStackID, stack, false)
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update environment status")
	}
}

func UpdatePausedEnvironmentStatus(targetVersion int, envStatus portainer.EdgeStackStatus) *portainer.EdgeStackStatus {
	if envStatus.DeploymentInfo.FileVersion >= targetVersion {
		return nil
	}

	if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
		return sts.Type == portainer.EdgeStackStatusPausedDeploying
	}) {
		// if the status is already set to paused, we can skip
		return nil
	}

	if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
		return sts.Type == portainer.EdgeStackStatusError
	}) {
		// if the status is already set to error, we can skip
		return nil
	}

	// if the status hasn't set to paused yet, we need to update it
	envStatus.Status = append(envStatus.Status, portainer.EdgeStackDeploymentStatus{
		Type:  portainer.EdgeStackStatusPausedDeploying,
		Error: "",
		Time:  time.Now().Unix(), // todo: maybe need to use last update time+1
	}, portainer.EdgeStackDeploymentStatus{
		Type:  portainer.EdgeStackStatusRunning,
		Error: "",
		Time:  time.Now().Unix(), // todo: maybe need to use last update time+1
	})

	return &envStatus
}

func UpdateRollingBackEnvironmentStatus(targetVersion int, envStatus portainer.EdgeStackStatus) *portainer.EdgeStackStatus {
	if envStatus.DeploymentInfo.FileVersion >= targetVersion {
		return nil
	}

	if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
		return sts.Type == portainer.EdgeStackStatusRolledBack
	}) {
		return nil
	}

	// If the current version is lower than the target version
	if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
		return sts.Type == portainer.EdgeStackStatusRunning
	}) {
		// if the status is running, it means that the stack was successfully
		// deployed in the endpoint, so it needs to be rolling back now.
		// we can add RollingBack status
		envStatus.Status = append(envStatus.Status, portainer.EdgeStackDeploymentStatus{
			Type:  portainer.EdgeStackStatusRollingBack,
			Error: "",
			Time:  time.Now().Unix(), // todo: maybe need to use last update time+1
		})

		return &envStatus
	}

	if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
		return sts.Type == portainer.EdgeStackStatusPending
	}) || len(envStatus.Status) == 0 {
		// if the status is pending, it mean that the stack has not been
		// deployed in the environment yet, so we can set the status to Running
		envStatus.Status = append(envStatus.Status, portainer.EdgeStackDeploymentStatus{
			Type:  portainer.EdgeStackStatusRunning,
			Error: "",
			Time:  time.Now().Unix(), // todo: maybe need to use last update time+1
		})

		return &envStatus
	}
	return nil
}
