package stacks

import (
	"log"
	"net/http"
	"time"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/scheduler"
	"github.com/portainer/portainer-ee/api/stacks"
)

func startAutoupdate(stackID portaineree.StackID, interval string, scheduler *scheduler.Scheduler, stackDeployer stacks.StackDeployer, datastore dataservices.DataStore, gitService portaineree.GitService, activityService portaineree.UserActivityService) (jobID string, e *httperror.HandlerError) {
	d, err := time.ParseDuration(interval)
	if err != nil {
		return "", &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Unable to parse stack's auto update interval", Err: err}
	}

	jobID = scheduler.StartJobEvery(d, func() error {
		return stacks.RedeployWhenChanged(stackID, stackDeployer, datastore, gitService, activityService)
	})

	return jobID, nil
}

func stopAutoupdate(stackID portaineree.StackID, jobID string, scheduler scheduler.Scheduler) {
	if jobID == "" {
		return
	}

	if err := scheduler.StopJob(jobID); err != nil {
		log.Printf("[WARN] could not stop the job for the stack %v", stackID)
	}

}
