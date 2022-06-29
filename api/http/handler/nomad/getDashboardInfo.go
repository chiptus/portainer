package nomad

import (
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/nomad/helpers"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
	"net/http"
)

type (
	DashboardPayload struct {
		NodeCount        int `json:"NodeCount"`
		JobCount         int `json:"JobCount"`
		GroupCount       int `json:"GroupCount"`
		TaskCount        int `json:"TaskCount"`
		RunningTaskCount int `json:"RunningTaskCount"`
	}
)

// @id getDashboard
// @summary get basic Nomad information for dashboard
// @description **Access policy**: authenticated users
// @tags nomad
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 "Success"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /nomad/dashboard [get]
func (handler *Handler) getDashboard(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid query parameter: endpointId", Err: err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to establish communication with Nomad server", Err: err}
	}

	dashboardPayload := DashboardPayload{}

	// node count
	nodeList, err := nomadClient.ListNodes()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to list nodes", Err: err}
	}
	dashboardPayload.NodeCount = len(nodeList)

	// job count
	jobList, err := nomadClient.ListJobs("*")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to list jobs", Err: err}
	}
	dashboardPayload.JobCount = len(jobList)

	// group and task count
	for _, job := range jobList {
		groups := job.JobSummary.Summary
		dashboardPayload.GroupCount += len(groups)

		for _, group := range groups {
			dashboardPayload.TaskCount += helpers.CalcGroupTasks(group)
			dashboardPayload.RunningTaskCount += group.Running
		}
	}

	return response.JSON(w, dashboardPayload)
}
