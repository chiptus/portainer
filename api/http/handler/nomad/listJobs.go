package nomad

import (
	"net/http"
	"time"

	"github.com/portainer/portainer-ee/api/http/middlewares"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type (
	TaskPayload struct {
		JobID        string    `json:"JobID"`
		Namespace    string    `json:"Namespace"`
		TaskName     string    `json:"TaskName"`
		State        string    `json:"State"`
		TaskGroup    string    `json:"TaskGroup"`
		AllocationID string    `json:"AllocationID"`
		StartedAt    time.Time `json:"StartedAt"`
	}

	JobPayload struct {
		ID         string        `json:"ID"`
		Status     string        `json:"Status"`
		Namespace  string        `json:"Namespace"`
		SubmitTime int64         `json:"SubmitTime"`
		Tasks      []TaskPayload `json:"Tasks"`
	}
)

// @id listJobs
// @summary List jobs
// @description namespace param is required
// @description **Access policy**: authenticated users
// @tags nomad
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param environmentId path int true "Environment identifier"
// @success 200 "Success"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /nomad/endpoints/{environmentId}/jobs [get]
func (handler *Handler) listJobs(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to establish communication with Nomad server", err)
	}

	jobList, err := nomadClient.ListJobs("*")
	if err != nil {
		return httperror.InternalServerError("Unable to list jobs", err)
	}

	jobsPayload := make([]JobPayload, 0)

	for _, job := range jobList {
		jobPayload := JobPayload{
			ID:         job.ID,
			Status:     job.Status,
			Namespace:  job.Namespace,
			SubmitTime: time.UnixMicro(job.SubmitTime).Unix(),
			Tasks:      []TaskPayload{},
		}

		allocations, err := nomadClient.ListAllocations(job.ID, job.Namespace)
		if err != nil {
			return httperror.InternalServerError("Unable to list allocations", err)
		}

		for _, allocation := range allocations {
			for taskName, taskState := range allocation.TaskStates {
				taskPayload := TaskPayload{
					JobID:        job.ID,
					Namespace:    job.Namespace,
					TaskName:     taskName,
					State:        taskState.State,
					TaskGroup:    allocation.TaskGroup,
					AllocationID: allocation.ID,
					StartedAt:    taskState.StartedAt,
				}
				jobPayload.Tasks = append(jobPayload.Tasks, taskPayload)
			}
		}

		jobsPayload = append(jobsPayload, jobPayload)
	}

	return response.JSON(w, jobsPayload)
}
