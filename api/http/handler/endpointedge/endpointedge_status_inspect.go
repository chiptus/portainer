package endpointedge

import (
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"

	"net/http"
	"strconv"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"

	"github.com/portainer/portainer-ee/api/http/middlewares"
)

type stackStatusResponse struct {
	// EdgeStack Identifier
	ID portaineree.EdgeStackID `example:"1"`
	// Version of this stack
	Version int `example:"3"`
}

type edgeJobResponse struct {
	// EdgeJob Identifier
	ID portaineree.EdgeJobID `json:"Id" example:"2"`
	// Whether to collect logs
	CollectLogs bool `json:"CollectLogs" example:"true"`
	// A cron expression to schedule this job
	CronExpression string `json:"CronExpression" example:"* * * * *"`
	// Script to run
	Script string `json:"Script" example:"echo hello"`
	// Version of this EdgeJob
	Version int `json:"Version" example:"2"`
}

type endpointEdgeStatusInspectResponse struct {
	// Status represents the environment(endpoint) status
	Status string `json:"status" example:"REQUIRED"`
	// The tunnel port
	Port int `json:"port" example:"8732"`
	// List of requests for jobs to run on the environment(endpoint)
	Schedules []edgeJobResponse `json:"schedules"`
	// The current value of CheckinInterval
	CheckinInterval int `json:"checkin" example:"5"`
	//
	Credentials string `json:"credentials"`
	// List of stacks to be deployed on the environments(endpoints)
	Stacks []stackStatusResponse `json:"stacks"`
}

// @id EndpointEdgeStatusInspect
// @summary Get environment(endpoint) status
// @description environment(endpoint) for edge agent to check status of environment(endpoint)
// @description **Access policy**: restricted only to Edge environments(endpoints)
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} endpointEdgeStatusInspectResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access environment(endpoint)"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/edge/status [get]
func (handler *Handler) endpointEdgeStatusInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	if endpoint.EdgeID == "" {
		edgeIdentifier := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)
		endpoint.EdgeID = edgeIdentifier
	}

	agentPlatform, agentPlatformErr := parseAgentPlatform(r)
	if agentPlatformErr != nil {
		return httperror.BadRequest("agent platform header is not valid", err)
	}
	endpoint.Type = agentPlatform

	timeZone := r.Header.Get(portaineree.PortainerAgentTimeZoneHeader)
	if timeZone != "" {
		endpoint.LocalTimeZone = timeZone
	}

	version := r.Header.Get(portaineree.PortainerAgentHeader)
	endpoint.Agent.Version = version

	endpoint.LastCheckInDate = time.Now().Unix()

	err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to Unable to persist environment changes inside the database", err)
	}

	checkinInterval := endpoint.EdgeCheckinInterval
	if endpoint.EdgeCheckinInterval == 0 {
		settings, err := handler.DataStore.Settings().Settings()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve settings from the database", err)
		}
		checkinInterval = settings.EdgeAgentCheckinInterval
	}

	tunnel := handler.ReverseTunnelService.GetTunnelDetails(endpoint.ID)

	statusResponse := endpointEdgeStatusInspectResponse{
		Status:          tunnel.Status,
		Port:            tunnel.Port,
		CheckinInterval: checkinInterval,
		Credentials:     tunnel.Credentials,
	}

	updateID, err := parseEdgeUpdateID(r)
	if err != nil {
		return httperror.BadRequest("Unable to parse edge update id", err)
	}

	// check endpoint version, if it has the same version as the active schedule, then we can mark the edge stack as successfully deployed
	activeUpdateSchedule := handler.edgeUpdateService.ActiveSchedule(endpoint.ID)
	if activeUpdateSchedule != nil && activeUpdateSchedule.ScheduleID == updateID {
		err := handler.handleSuccessfulUpdate(activeUpdateSchedule)
		if err != nil {
			return httperror.InternalServerError("Unable to handle successful update", err)
		}
	}

	schedules, handlerErr := handler.buildSchedules(endpoint.ID, tunnel)
	if handlerErr != nil {
		return handlerErr
	}
	statusResponse.Schedules = schedules

	if tunnel.Status == portaineree.EdgeAgentManagementRequired {
		handler.ReverseTunnelService.SetTunnelStatusToActive(endpoint.ID)
	}

	location, err := parseLocation(endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to parse location", err)
	}

	edgeStacksStatus, handlerErr := handler.buildEdgeStacks(endpoint.ID, location)
	if handlerErr != nil {
		return handlerErr
	}
	statusResponse.Stacks = edgeStacksStatus

	return response.JSON(w, statusResponse)
}

func parseLocation(endpoint *portaineree.Endpoint) (*time.Location, error) {
	if endpoint.LocalTimeZone == "" {
		return nil, nil
	}

	location, err := time.LoadLocation(endpoint.LocalTimeZone)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load location")
	}

	return location, nil
}

func parseAgentPlatform(r *http.Request) (portaineree.EndpointType, error) {
	agentPlatformHeader := r.Header.Get(portaineree.HTTPResponseAgentPlatform)
	if agentPlatformHeader == "" {
		return 0, errors.New("agent platform header is missing")
	}

	agentPlatformNumber, err := strconv.Atoi(agentPlatformHeader)
	if err != nil {
		return 0, err
	}

	agentPlatform := portaineree.AgentPlatform(agentPlatformNumber)

	switch agentPlatform {
	case portaineree.AgentPlatformDocker:
		return portaineree.EdgeAgentOnDockerEnvironment, nil
	case portaineree.AgentPlatformKubernetes:
		return portaineree.EdgeAgentOnKubernetesEnvironment, nil
	case portaineree.AgentPlatformNomad:
		return portaineree.EdgeAgentOnNomadEnvironment, nil
	default:
		return 0, fmt.Errorf("agent platform %v is not valid", agentPlatform)
	}
}

func (handler *Handler) updateEdgeStackStatus(edgeStack *portaineree.EdgeStack, environmentID portaineree.EndpointID) error {
	status, ok := edgeStack.Status[environmentID]
	if !ok {
		status = portaineree.EdgeStackStatus{
			EndpointID: environmentID,
		}
	}

	status.Type = portaineree.EdgeStackStatusRemoteUpdateSuccess
	status.Error = ""

	edgeStack.Status[environmentID] = status
	return handler.DataStore.EdgeStack().UpdateEdgeStack(edgeStack.ID, edgeStack)
}

func (handler *Handler) handleSuccessfulUpdate(activeUpdateSchedule *edgetypes.EndpointUpdateScheduleRelation) error {
	handler.edgeUpdateService.RemoveActiveSchedule(activeUpdateSchedule.EnvironmentID, activeUpdateSchedule.ScheduleID)

	edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(activeUpdateSchedule.EdgeStackID)
	if err != nil {
		return err
	}

	return handler.updateEdgeStackStatus(edgeStack, activeUpdateSchedule.EnvironmentID)
}

func (handler *Handler) buildSchedules(endpointID portaineree.EndpointID, tunnel portaineree.TunnelDetails) ([]edgeJobResponse, *httperror.HandlerError) {
	schedules := []edgeJobResponse{}
	for _, job := range tunnel.Jobs {
		schedule := edgeJobResponse{
			ID:             job.ID,
			CronExpression: job.CronExpression,
			CollectLogs:    job.Endpoints[endpointID].CollectLogs,
			Version:        job.Version,
		}

		file, err := handler.FileService.GetFileContent(job.ScriptPath, "")
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve Edge job script file", err)
		}
		schedule.Script = base64.RawStdEncoding.EncodeToString(file)

		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (handler *Handler) buildEdgeStacks(endpointID portaineree.EndpointID, timeZone *time.Location) ([]stackStatusResponse, *httperror.HandlerError) {
	relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpointID)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve relation object from the database", err)
	}

	edgeStacksStatus := []stackStatusResponse{}
	for stackID := range relation.EdgeStacks {
		stack, err := handler.DataStore.EdgeStack().EdgeStack(stackID)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve edge stack from the database", err)
		}

		// if the stack represents a successful remote update or failed - skip it
		if endpointStatus, ok := stack.Status[endpointID]; ok && (endpointStatus.Type == portaineree.EdgeStackStatusRemoteUpdateSuccess || (stack.EdgeUpdateID != 0 && endpointStatus.Type == portaineree.StatusError)) {
			continue
		}

		pastSchedule, err := shouldScheduleTrigger(stack.ScheduledTime, timeZone)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to parse scheduled time", err)
		}

		if !pastSchedule {
			continue
		}

		stackStatus := stackStatusResponse{
			ID:      stack.ID,
			Version: stack.Version,
		}

		edgeStacksStatus = append(edgeStacksStatus, stackStatus)
	}
	return edgeStacksStatus, nil
}

func shouldScheduleTrigger(scheduledTime string, location *time.Location) (bool, error) {
	if location == nil || scheduledTime == "" {
		return true, nil
	}

	localScheduledTime, err := time.ParseInLocation(portaineree.DateTimeFormat, scheduledTime, location)
	if err != nil {
		return false, errors.WithMessage(err, "unable to parse scheduled time")
	}

	localTime := time.Now().In(location)
	return localScheduledTime.Before(localTime), nil
}

func parseEdgeUpdateID(r *http.Request) (edgetypes.UpdateScheduleID, error) {
	value := r.Header.Get(portaineree.PortainerAgentEdgeUpdateIDHeader)
	if value == "" {
		return 0, nil
	}

	updateID, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.WithMessage(err, "unable to parse edge update ID")
	}

	return edgetypes.UpdateScheduleID(updateID), nil
}
