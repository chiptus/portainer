package deployments

import (
	"time"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/scheduler"

	"github.com/rs/zerolog/log"
)

func StartAutoupdate(stackID portaineree.StackID, interval string, scheduler *scheduler.Scheduler, stackDeployer StackDeployer, datastore dataservices.DataStore, gitService portaineree.GitService, activityService portaineree.UserActivityService) (jobID string, e *httperror.HandlerError) {
	d, err := time.ParseDuration(interval)
	if err != nil {
		return "", httperror.BadRequest("Unable to parse stack's auto update interval", err)
	}

	jobID = scheduler.StartJobEvery(d, func() error {
		return RedeployWhenChanged(stackID, stackDeployer, datastore, gitService, activityService, nil)
	})

	return jobID, nil
}

func StopAutoupdate(stackID portaineree.StackID, jobID string, scheduler *scheduler.Scheduler) {
	if jobID == "" {
		return
	}

	if err := scheduler.StopJob(jobID); err != nil {
		log.Warn().Int("stack_id", int(stackID)).Msg("could not stop the job for the stack")
	}
}
