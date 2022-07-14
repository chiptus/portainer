package nomad

import (
	"net/http"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/http/middlewares"
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
// @success 200 "Success"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /nomad/endpoints/{endpointID}/jobs [get]
func (handler *Handler) listJobs(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to establish communication with Nomad server", Err: err}
	}

	jobList, err := nomadClient.ListJobs("*")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to list jobs", Err: err}
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
			return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to list allocations", Err: err}
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
