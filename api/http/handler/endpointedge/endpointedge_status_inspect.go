package endpointedge

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/portainer/portainer-ee/api/internal/slices"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/pkg/featureflags"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	//
	EdgeConfigurations map[portaineree.EdgeConfigID]portaineree.EdgeConfigStateType `json:"edge_configurations"`
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
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	cachedResp := handler.respondFromCache(w, r, portaineree.EndpointID(endpointID))
	if cachedResp {
		return nil
	}

	if _, ok := handler.DataStore.Endpoint().Heartbeat(portaineree.EndpointID(endpointID)); !ok {
		// EE-5190
		return httperror.Forbidden("Permission denied to access environment", errors.New("the device has not been trusted yet"))
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err != nil {
		// EE-5190
		return httperror.Forbidden("Permission denied to access environment", errors.New("the device has not been trusted yet"))
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	handler.DataStore.Endpoint().UpdateHeartbeat(endpoint.ID)

	err = handler.requestBouncer.TrustedEdgeEnvironmentAccess(handler.DataStore, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	var statusResponse *endpointEdgeStatusInspectResponse
	var skipCache bool
	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		statusResponse, skipCache, err = handler.inspectStatus(handler.DataStore, r, endpoint.ID)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			statusResponse, skipCache, err = handler.inspectStatus(tx, r, endpoint.ID)
			return err
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return cacheResponse(w, endpoint.ID, *statusResponse, skipCache)
}

func (handler *Handler) inspectStatus(tx dataservices.DataStoreTx, r *http.Request, endpointID portaineree.EndpointID) (*endpointEdgeStatusInspectResponse, bool, error) {
	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if err != nil {
		return nil, false, err
	}

	if endpoint.EdgeID == "" {
		edgeIdentifier := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)
		endpoint.EdgeID = edgeIdentifier
	}

	agentPlatform, agentPlatformErr := parseAgentPlatform(r)
	if agentPlatformErr != nil {
		return nil, false, httperror.BadRequest("agent platform header is not valid", agentPlatformErr)
	}
	endpoint.Type = agentPlatform

	timeZone := r.Header.Get(portaineree.PortainerAgentTimeZoneHeader)
	if timeZone != "" {
		endpoint.LocalTimeZone = timeZone
	}

	version := r.Header.Get(portaineree.PortainerAgentHeader)
	endpoint.Agent.Version = version

	// Take an initial snapshot
	if endpoint.LastCheckInDate == 0 {
		handler.ReverseTunnelService.SetTunnelStatusToRequired(endpoint.ID)
	}

	endpoint.LastCheckInDate = time.Now().Unix()

	err = tx.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return nil, false, httperror.InternalServerError("Unable to persist environment changes inside the database", err)
	}

	checkinInterval := endpoint.EdgeCheckinInterval
	if endpoint.EdgeCheckinInterval == 0 {
		settings, err := tx.Settings().Settings()
		if err != nil {
			return nil, false, httperror.InternalServerError("Unable to retrieve settings from the database", err)
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
		return nil, false, httperror.BadRequest("Unable to parse edge update id", err)
	}

	// check endpoint version, if it has the same version as the active schedule, then we can mark the edge stack as successfully deployed
	activeUpdateSchedule := handler.edgeUpdateService.ActiveSchedule(endpoint.ID)
	if activeUpdateSchedule != nil && activeUpdateSchedule.ScheduleID == updateID {
		err := handler.handleSuccessfulUpdate(tx, activeUpdateSchedule)
		if err != nil {
			return nil, false, httperror.InternalServerError("Unable to handle successful update", err)
		}
	}

	schedules, handlerErr := handler.buildSchedules(endpoint.ID, tunnel)
	if handlerErr != nil {
		return nil, false, handlerErr
	}
	statusResponse.Schedules = schedules

	if tunnel.Status == portaineree.EdgeAgentManagementRequired {
		handler.ReverseTunnelService.SetTunnelStatusToActive(endpoint.ID)
	}

	location, err := parseLocation(endpoint)
	if err != nil {
		return nil, false, httperror.InternalServerError("Unable to parse location", err)
	}

	skipCache := false

	if updateID > 0 || activeUpdateSchedule != nil {
		// To determine if a request comes from an agent after an update, we
		// rely on the condition that updateID > 0. However, in addition to that,
		// we also need to check whether the request is intended to update an agent.
		// To do so, we verify that activeUpdateSchedule is not nil.
		skipCache = true
	}
	edgeStacksStatus, handlerErr := handler.buildEdgeStacks(tx, endpoint.ID, location, &skipCache)
	if handlerErr != nil {
		return nil, skipCache, handlerErr
	}
	statusResponse.Stacks = edgeStacksStatus

	configState, err := handler.DataStore.EdgeConfigState().Read(endpoint.ID)
	if err != nil && !handler.DataStore.IsErrObjectNotFound(err) {
		return nil, skipCache, httperror.InternalServerError("Unable to retrieve edge config state from the database", err)
	} else if err == nil {
		for edgeConfigID, state := range configState.States {
			switch state {
			case portaineree.EdgeConfigIdleState, portaineree.EdgeConfigFailureState:
				continue
			}

			if statusResponse.EdgeConfigurations == nil {
				statusResponse.EdgeConfigurations = make(map[portaineree.EdgeConfigID]portaineree.EdgeConfigStateType)
			}

			statusResponse.EdgeConfigurations[edgeConfigID] = state
		}
	}

	return &statusResponse, skipCache, nil
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

func (handler *Handler) updateEdgeStackStatus(tx dataservices.DataStoreTx, edgeStack *portaineree.EdgeStack, environmentID portaineree.EndpointID) error {
	status, ok := edgeStack.Status[environmentID]
	if !ok {
		status = portainer.EdgeStackStatus{
			Status:         []portainer.EdgeStackDeploymentStatus{},
			EndpointID:     portainer.EndpointID(environmentID),
			DeploymentInfo: portainer.StackDeploymentInfo{},
		}
	}

	status.Status = append(status.Status, portainer.EdgeStackDeploymentStatus{
		Type: portainer.EdgeStackStatusRemoteUpdateSuccess,
		Time: time.Now().Unix(),
	})

	edgeStack.Status[environmentID] = status

	return tx.EdgeStack().UpdateEdgeStack(edgeStack.ID, edgeStack, true)
}

func (handler *Handler) handleSuccessfulUpdate(tx dataservices.DataStoreTx, activeUpdateSchedule *edgetypes.EndpointUpdateScheduleRelation) error {
	handler.edgeUpdateService.RemoveActiveSchedule(activeUpdateSchedule.EnvironmentID, activeUpdateSchedule.ScheduleID)

	edgeStack, err := tx.EdgeStack().EdgeStack(activeUpdateSchedule.EdgeStackID)
	if err != nil {
		return err
	}

	return handler.updateEdgeStackStatus(tx, edgeStack, activeUpdateSchedule.EnvironmentID)
}

func (handler *Handler) buildSchedules(endpointID portaineree.EndpointID, tunnel portaineree.TunnelDetails) ([]edgeJobResponse, *httperror.HandlerError) {
	schedules := []edgeJobResponse{}
	for _, job := range tunnel.Jobs {
		var collectLogs bool
		if _, ok := job.GroupLogsCollection[endpointID]; ok {
			collectLogs = job.GroupLogsCollection[endpointID].CollectLogs
		} else {
			collectLogs = job.Endpoints[endpointID].CollectLogs
		}

		schedule := edgeJobResponse{
			ID:             job.ID,
			CronExpression: job.CronExpression,
			CollectLogs:    collectLogs,
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

func (handler *Handler) buildEdgeStacks(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, timeZone *time.Location, skipCache *bool) ([]stackStatusResponse, *httperror.HandlerError) {
	relation, err := tx.EndpointRelation().EndpointRelation(endpointID)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve relation object from the database", err)
	}

	edgeStacksStatus := []stackStatusResponse{}
	for stackID := range relation.EdgeStacks {
		version, ok := tx.EdgeStack().EdgeStackVersion(stackID)
		if !ok {
			return nil, httperror.InternalServerError("Unable to retrieve edge stack from the database", err)
		}

		var stack *portaineree.EdgeStack

		if skipCache != nil && *skipCache {
			// If the edge stack is intended for the updater, there is a potential issue with the cachedStack.
			// For instance, if a group of 5 agents is scheduled for an update and all 5 agents are updated successfully,
			// the first newly added agent will query the "/endpoints/{id}/edge/status" API endpoint, as specified in this
			// file, and set the corresponding RemoteUpdateSuccess value in EdgeStack.Status to true. The updated edge
			// stack copy is then added to the edgeStackCache, with an expiration time of 5 seconds. If another new agent
			// spins up and queries the API within the 5-second window, its RemoteUpdateSuccess value will also be updated
			// to "true," but instead of using the new value, its previous value of "false" stored in the edgeStackCache will
			// be used in the API call. As a result, the new agent will deploy the update schedule for the new agent again,
			// leading to a chain of incorrect behavior.
			stack, err = tx.EdgeStack().EdgeStack(stackID)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to retrieve an edge stack from the database", err)
			}
		} else {
			cacheKey := strconv.Itoa(int(stackID))
			cachedStack, ok := handler.edgeStackCache.Get(cacheKey)
			if ok {
				stack, ok = cachedStack.(*portaineree.EdgeStack)
				if !ok {
					return nil, httperror.InternalServerError("", errors.New(""))
				}
			} else {
				stack, err = tx.EdgeStack().EdgeStack(stackID)
				if err != nil {
					return nil, httperror.InternalServerError("Unable to retrieve an edge stack from the database", err)
				}

				_ = handler.edgeStackCache.Add(cacheKey, stack, portaineree.DefaultEdgeAgentCheckinIntervalInSeconds*time.Second)
			}
		}

		if stack.EdgeUpdateID != 0 {
			endpointStatus, ok := stack.Status[endpointID]
			if !ok {
				endpointStatus = portainer.EdgeStackStatus{
					Status:     []portainer.EdgeStackDeploymentStatus{},
					EndpointID: portainer.EndpointID(endpointID),
				}
			}

			// if the stack represents a successful remote update or failed - skip it
			if slices.ContainsFunc(endpointStatus.Status, func(s portainer.EdgeStackDeploymentStatus) bool {
				return s.Type == portainer.EdgeStackStatusRemoteUpdateSuccess || s.Type == portainer.EdgeStackStatusError
			}) {
				continue
			}
		}

		pastSchedule, err := shouldScheduleTrigger(stack.ScheduledTime, timeZone)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to parse scheduled time", err)
		}

		if !pastSchedule {
			continue
		}

		stackStatus := stackStatusResponse{
			ID:      stackID,
			Version: version,
		}

		// Stagger configuration check
		if handler.staggerService != nil && handler.staggerService.IsStaggeredEdgeStack(stackID, stack.StackFileVersion, endpointID) {
			log.Debug().Int("edgeStackID", int(stackID)).
				Int("endpointID", int(endpointID)).
				Int("statusVersion", stackStatus.Version).
				Int("fileVersion", stack.StackFileVersion).
				Msg("[stagger status inspect] Staggered edge stack")

			if handler.staggerService.MarkedAsCompleted(stackID, stack.StackFileVersion) {
				log.Debug().Msg("[stagger status inspect] Marked as completed, skip")

				// There is a chance that endpoint status has updated in the databse, but the
				// cached stack is still the previous one. So we need to retrieve the latest
				// stack from the database
				edgeStack, err := tx.EdgeStack().EdgeStack(stackID)
				if err == nil {
					if edgeStack.StackFileVersion != stack.StackFileVersion {
						// If the stack file version is not the latest one, we need to skip the cache
						// ETag for the current status request.
						// This case can happen when the user quickly redeploy a new stagger update right
						// after the previous stagger update is completed. In the middle of the update, the
						// agent polls the status of the edge stack with the previous file version. This is
						// an edge case
						*skipCache = true
					}
					stack = edgeStack
				}

				endpointStatus, ok := stack.Status[endpointID]
				if !ok {
					// If the endpoint was updated in the stagger flow, and then removed from the
					// edge group and added back, the endpoint status will be removed from the
					// stack. In this case, we need to treat it as a new endpoint
					edgeStacksStatus = append(edgeStacksStatus, stackStatus)
					continue
				}
				if endpointStatus.DeploymentInfo.Version == 0 {
					// The endpoint has never deployed an edge stack, such as, a failure endpoint
					continue
				}

				// if the stagger update is completed, we should always return the
				// version that the endpoint has deployed
				stackStatus.Version = endpointStatus.DeploymentInfo.Version
				edgeStacksStatus = append(edgeStacksStatus, stackStatus)
				continue
			}

			// If the edge stack is staggered, check if the endpoint is in the current stagger queue
			if !handler.staggerService.CanProceedAsStaggerJob(stackID, stack.StackFileVersion, endpointID) {
				// It's not the turn for the endpoint, skip

				// skip the cache, otherwise the edge agent will not be able to get the stack status when
				// it's the turn for the stagger queue
				*skipCache = true

				// prevent waiting edge agents from being removed before they got processed
				endpointStatus := stack.Status[endpointID]
				if endpointStatus.DeploymentInfo.Version == 0 {
					continue
				}

				log.Debug().
					Int("endpointID", int(endpointID)).
					Int("endpointVersion", endpointStatus.DeploymentInfo.Version).
					Msg("[stagger status inspect] Cannot proceed as stagger job")

				stackStatus.Version = endpointStatus.DeploymentInfo.Version
				edgeStacksStatus = append(edgeStacksStatus, stackStatus)
				continue
			}

			log.Debug().Msg("[stagger status inspect] Can proceed as stagger job")

			// If the deployed version for the endpoint is already rolled back, skip
			if handler.staggerService.MarkedAsRollback(stackID, stack.StackFileVersion) {
				if handler.staggerService.WasEndpointRolledBack(stackID, stack.StackFileVersion, endpointID) {
					edgeStack, err := tx.EdgeStack().EdgeStack(stackID)
					if err == nil {
						stack = edgeStack
					}
					// if the endpoint was already rolled back, we need to return its current deployed version
					endpointStatus := stack.Status[endpointID]
					log.Debug().Int("endpointID", int(endpointID)).
						Int("endpointVersion", endpointStatus.DeploymentInfo.Version).
						Msg("[stagger status inspect] was endpoint rolled back")

					// if the endpoint was not updated yet, the version will be the stack version
					// if the endpoint was updated and rolled back, the version will be
					// the previous stack file version, which is different from stack version
					stackStatus.Version = endpointStatus.DeploymentInfo.Version
					edgeStacksStatus = append(edgeStacksStatus, stackStatus)
					continue
				}

				*skipCache = true
				// todo: if we support to store more versions of the stack file in the future,
				// we can support to rollback to previous version against each endpoint
				if stack.PreviousDeploymentInfo != nil {
					// if the endpoint was not rolled back yet, we need to return the previous version
					stackStatus.Version = stack.PreviousDeploymentInfo.FileVersion
				} else {
					log.Warn().Msg("No previous deployment info found")
				}
			}
		}

		edgeStacksStatus = append(edgeStacksStatus, stackStatus)
	}

	return edgeStacksStatus, nil
}

func cacheResponse(w http.ResponseWriter, endpointID portaineree.EndpointID, statusResponse endpointEdgeStatusInspectResponse, skipCache bool) *httperror.HandlerError {
	rr := httptest.NewRecorder()

	httpErr := response.JSON(rr, statusResponse)
	if httpErr != nil {
		return httpErr
	}

	resp := rr.Result()

	if !skipCache {
		h := fnv.New32a()
		h.Write(rr.Body.Bytes())
		etag := strconv.FormatUint(uint64(h.Sum32()), 16)

		cache.Set(endpointID, []byte(etag))

		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}

		w.Header().Set("ETag", etag)
	}

	io.Copy(w, resp.Body)

	return nil
}

func (handler *Handler) respondFromCache(w http.ResponseWriter, r *http.Request, endpointID portaineree.EndpointID) bool {
	inmHeader := r.Header.Get("If-None-Match")
	etags := strings.Split(inmHeader, ",")

	if len(inmHeader) == 0 || etags[0] == "" {
		return false
	}

	cachedETag, ok := cache.Get(endpointID)
	if !ok {
		return false
	}

	for _, etag := range etags {
		if !bytes.Equal([]byte(etag), cachedETag) {
			continue
		}

		handler.DataStore.Endpoint().UpdateHeartbeat(endpointID)

		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusNotModified)

		return true
	}

	return false
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
