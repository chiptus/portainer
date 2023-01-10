package endpointedge

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type EdgeAsyncRequest struct {
	CommandTimestamp *time.Time             `json:"commandTimestamp"`
	Snapshot         *snapshot              `json:"snapshot"`
	EndpointID       portaineree.EndpointID `json:"endpointId"`
}

const (
	edgeIntervalUseDefault = -1
)

type snapshot struct {
	Docker      *portainer.DockerSnapshot `json:"docker,omitempty"`
	DockerPatch jsonpatch.Patch           `json:"dockerPatch,omitempty"`
	DockerHash  *uint32                   `json:"dockerHash,omitempty"`

	Kubernetes      *portaineree.KubernetesSnapshot `json:"kubernetes,omitempty"`
	KubernetesPatch jsonpatch.Patch                 `json:"kubernetesPatch,omitempty"`
	KubernetesHash  *uint32                         `json:"kubernetesHash,omitempty"`

	StackLogs   []portaineree.EdgeStackLog                            `json:"stackLogs,omitempty"`
	StackStatus map[portaineree.EdgeStackID]portainer.EdgeStackStatus `json:"stackStatus,omitempty"`
	JobsStatus  map[portaineree.EdgeJobID]portaineree.EdgeJobStatus   `json:"jobsStatus:,omitempty"`
}

func (payload *EdgeAsyncRequest) Validate(r *http.Request) error {
	return nil
}

type EdgeAsyncResponse struct {
	EndpointID portaineree.EndpointID `json:"endpointID"`

	PingInterval     time.Duration `json:"pingInterval"`
	SnapshotInterval time.Duration `json:"snapshotInterval"`
	CommandInterval  time.Duration `json:"commandInterval"`

	Commands         []portaineree.EdgeAsyncCommand `json:"commands"`
	NeedFullSnapshot bool                           `json:"needFullSnapshot"`
}

// @id endpointEdgeAsync
// @summary Get environment(endpoint) status
// @description Environment(Endpoint) for edge agent to check status of environment(endpoint)
// @description **Access policy**: restricted only to Edge environments(endpoints)
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} EdgeAsyncResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access environment(endpoint)"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/edge/async [post]
func (handler *Handler) endpointEdgeAsync(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var err error

	edgeID := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)
	if edgeID == "" {
		log.Debug().Str("PortainerAgentEdgeIDHeader", edgeID).Msg("missing agent edge id")

		return httperror.BadRequest("missing Edge identifier", errors.New("missing Edge identifier"))
	}

	payload, err := parseBodyPayload(r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	endpoint, err := handler.getEndpoint(payload.EndpointID, edgeID)
	if err != nil {
		return httperror.InternalServerError("Endpoint with edge id or endpoint id is missing", err)
	}

	version := r.Header.Get(portaineree.PortainerAgentHeader)

	timeZone := r.Header.Get(portaineree.PortainerAgentTimeZoneHeader)

	if endpoint == nil {
		log.Debug().Str("PortainerAgentEdgeIDHeader", edgeID).Msg("edge id not found in existing endpoints")

		agentPlatform, agentPlatformErr := parseAgentPlatform(r)
		if agentPlatformErr != nil {
			return httperror.BadRequest("agent platform header is not valid", err)
		}

		validateCertsErr := handler.requestBouncer.AuthorizedClientTLSConn(r)
		if validateCertsErr != nil {
			return httperror.Forbidden("Permission denied to access environment", err)
		}

		newEndpoint, createEndpointErr := handler.createAsyncEdgeAgentEndpoint(r, edgeID, agentPlatform, version, timeZone)
		if createEndpointErr != nil {
			return createEndpointErr
		}

		asyncResponse := EdgeAsyncResponse{
			EndpointID: newEndpoint.ID,
		}

		return response.JSON(w, asyncResponse)
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
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

		err = handler.EdgeService.RemoveStackCommand(endpoint.ID, activeUpdateSchedule.EdgeStackID)
		if err != nil {
			return httperror.InternalServerError("Unable to replace stack command", err)
		}
	}

	var needFullSnapshot bool
	if payload.Snapshot != nil {
		handler.saveSnapshot(endpoint, payload.Snapshot, &needFullSnapshot)
	}

	if timeZone != "" {
		endpoint.LocalTimeZone = timeZone
	}
	endpoint.LastCheckInDate = time.Now().Unix()
	endpoint.Status = portaineree.EndpointStatusUp
	endpoint.Edge.AsyncMode = true
	endpoint.Agent.Version = version

	err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to Unable to persist environment changes inside the database", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	pingInterval := endpoint.Edge.PingInterval
	if pingInterval == edgeIntervalUseDefault {
		pingInterval = settings.Edge.PingInterval
	}

	snapshotInterval := endpoint.Edge.SnapshotInterval
	if snapshotInterval == edgeIntervalUseDefault {
		snapshotInterval = settings.Edge.SnapshotInterval
	}

	commandInterval := endpoint.Edge.CommandInterval
	if commandInterval == edgeIntervalUseDefault {
		commandInterval = settings.Edge.CommandInterval
	}

	asyncResponse := EdgeAsyncResponse{
		EndpointID:       endpoint.ID,
		PingInterval:     time.Duration(pingInterval) * time.Second,
		SnapshotInterval: time.Duration(snapshotInterval) * time.Second,
		CommandInterval:  time.Duration(commandInterval) * time.Second,
		NeedFullSnapshot: needFullSnapshot,
	}

	if payload.CommandTimestamp != nil {
		location, err := parseLocation(endpoint)
		if err != nil {
			return httperror.InternalServerError("Unable to parse location", err)
		}

		commands, err := handler.sendCommandsSince(endpoint, *payload.CommandTimestamp, location)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve commands", err)
		}

		asyncResponse.Commands = commands
	}

	return response.JSON(w, asyncResponse)
}

func parseBodyPayload(req *http.Request) (EdgeAsyncRequest, error) {
	var err error

	if req.Header.Get("Content-Encoding") == "gzip" {
		gzr, err := gzip.NewReader(req.Body)
		if err != nil {
			return EdgeAsyncRequest{}, errors.WithMessage(err, "Unable to read gzip body")
		}

		req.Body = gzr
	}

	var payload EdgeAsyncRequest
	err = request.DecodeAndValidateJSONPayload(req, &payload)
	if err != nil {
		log.Error().Err(err).Str("payload", fmt.Sprintf("%+v", req)).Msg("decode payload")
		return EdgeAsyncRequest{}, errors.WithMessage(err, "Unable to decode payload")
	}

	return payload, nil
}

func (handler *Handler) createAsyncEdgeAgentEndpoint(req *http.Request, edgeID string, endpointType portaineree.EndpointType, version string, timeZone string) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve the settings", err)
	}

	if settings.EdgePortainerURL == "" {
		return nil, httperror.InternalServerError("Portainer API server URL is not set in Edge Compute settings", errors.New("Portainer API server URL is not set in Edge Compute settings"))
	}

	if settings.Edge.TunnelServerAddress == "" {
		return nil, httperror.InternalServerError("Tunnel server address is not set in Edge Compute settings", errors.New("Tunnel server address is not set in Edge Compute settings"))
	}

	edgeKey := handler.ReverseTunnelService.GenerateEdgeKey(settings.EdgePortainerURL, settings.Edge.TunnelServerAddress, endpointID)

	endpoint := &portaineree.Endpoint{
		ID:     portaineree.EndpointID(endpointID),
		EdgeID: edgeID,
		Name:   edgeID,
		Type:   endpointType,
		TLSConfig: portaineree.TLSConfiguration{
			TLS: false,
		},
		GroupID:            portaineree.EndpointGroupID(1),
		AuthorizedUsers:    []portaineree.UserID{},
		AuthorizedTeams:    []portaineree.TeamID{},
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             []portaineree.TagID{},
		Status:             portaineree.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		EdgeKey:            edgeKey,
		Kubernetes:         portaineree.KubernetesDefault(),
		IsEdgeDevice:       true,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		LocalTimeZone: timeZone,
		SecuritySettings: portaineree.EndpointSecuritySettings{
			AllowVolumeBrowserForRegularUsers:         false,
			EnableHostManagementFeatures:              false,
			AllowSysctlSettingForRegularUsers:         true,
			AllowBindMountsForRegularUsers:            true,
			AllowPrivilegedModeForRegularUsers:        true,
			AllowHostNamespaceForRegularUsers:         true,
			AllowContainerCapabilitiesForRegularUsers: true,
			AllowDeviceMappingForRegularUsers:         true,
			AllowStackManagementForRegularUsers:       true,
		},
	}

	endpoint.Agent.Version = version

	endpoint.Edge.AsyncMode = true
	endpoint.Edge.PingInterval = settings.Edge.PingInterval
	endpoint.Edge.SnapshotInterval = settings.Edge.SnapshotInterval
	endpoint.Edge.CommandInterval = settings.Edge.CommandInterval
	endpoint.UserTrusted = settings.TrustOnFirstConnect

	err = handler.DataStore.Endpoint().Create(endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	relationObject := &portaineree.EndpointRelation{
		EndpointID: endpoint.ID,
		EdgeStacks: map[portaineree.EdgeStackID]bool{},
	}

	err = handler.DataStore.EndpointRelation().Create(relationObject)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to persist the relation object inside the database", err)
	}

	return endpoint, nil
}

func (handler *Handler) updateDockerSnapshot(endpoint *portaineree.Endpoint, snapshotPayload *snapshot, needFullSnapshot *bool) error {
	snapshot := &portaineree.Snapshot{}

	if len(snapshotPayload.DockerPatch) > 0 {
		var err error
		snapshot, err = handler.DataStore.Snapshot().Snapshot(endpoint.ID)
		if err != nil || snapshot.Docker == nil {
			*needFullSnapshot = true
			return errors.New("received a Docker snapshot patch but there was no previous snapshot")
		}

		lastSnapshot := snapshot.Docker
		lastSnapshotJSON, err := json.Marshal(lastSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not marshal the last Docker snapshot: %s", err)
		}

		if snapshotPayload.DockerHash == nil {
			*needFullSnapshot = true
			return fmt.Errorf("the snapshot hash is missing")
		}

		h := snapshotHash(lastSnapshotJSON)
		if *snapshotPayload.DockerHash != h {
			*needFullSnapshot = true
			log.Debug().Uint32("expected", h).Uint32("got", *snapshotPayload.DockerHash).Msg("hash mismatch")

			return fmt.Errorf("the snapshot hash does not match against the stored one")
		}

		newSnapshotJSON, err := snapshotPayload.DockerPatch.Apply(lastSnapshotJSON)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not apply the patch to the Docker snapshot: %s", err)
		}

		var newSnapshot portainer.DockerSnapshot
		err = json.Unmarshal(newSnapshotJSON, &newSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not unmarshal the patched Docker snapshot: %s", err)
		}

		log.Debug().Msg("setting the patched Docker snapshot")
		snapshot.Docker = &newSnapshot
	} else if snapshotPayload.Docker != nil {
		log.Debug().Msg("setting the full Docker snapshot")
		snapshot.Docker = snapshotPayload.Docker
	}

	snapshot.EndpointID = endpoint.ID
	err := handler.DataStore.Snapshot().UpdateSnapshot(snapshot)
	if err != nil {
		return errors.New("snapshot could not be updated")
	}

	return nil
}

func (handler *Handler) updateKubernetesSnapshot(endpoint *portaineree.Endpoint, snapshotPayload *snapshot, needFullSnapshot *bool) error {
	snapshot := &portaineree.Snapshot{}

	if len(snapshotPayload.KubernetesPatch) > 0 {
		var err error
		snapshot, err = handler.DataStore.Snapshot().Snapshot(endpoint.ID)
		if err != nil || snapshot.Kubernetes == nil {
			*needFullSnapshot = true
			return errors.New("received a Kubernetes snapshot patch but there was no previous snapshot")
		}

		lastSnapshot := snapshot.Kubernetes
		lastSnapshotJSON, err := json.Marshal(lastSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not marshal the last Kubernetes snapshot: %s", err)
		}

		if snapshotPayload.KubernetesHash == nil {
			*needFullSnapshot = true
			return fmt.Errorf("the snapshot hash is missing")
		}

		h := snapshotHash(lastSnapshotJSON)
		if *snapshotPayload.KubernetesHash != h {
			*needFullSnapshot = true
			log.Debug().Uint32("expected", h).Uint32("got", *snapshotPayload.KubernetesHash).Msg("hash mismatch")

			return fmt.Errorf("the snapshot hash does not match against the stored one")
		}

		newSnapshotJSON, err := snapshotPayload.KubernetesPatch.Apply(lastSnapshotJSON)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not apply the patch to the Kubernetes snapshot: %s", err)
		}

		var newSnapshot portaineree.KubernetesSnapshot
		err = json.Unmarshal(newSnapshotJSON, &newSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not unmarshal the patched Kubernetes snapshot: %s", err)
		}

		log.Debug().Msg("setting the patched Kubernetes snapshot")
		snapshot.Kubernetes = &newSnapshot
	} else if snapshotPayload.Kubernetes != nil {
		log.Debug().Msg("setting the full Kubernetes snapshot")
		snapshot.Kubernetes = snapshotPayload.Kubernetes
	}

	snapshot.EndpointID = endpoint.ID
	err := handler.DataStore.Snapshot().UpdateSnapshot(snapshot)
	if err != nil {
		return errors.New("snapshot could not be updated")
	}

	return nil
}

func (handler *Handler) saveSnapshot(endpoint *portaineree.Endpoint, snapshotPayload *snapshot, needFullSnapshot *bool) {
	// Save edge stacks status
	for stackID, status := range snapshotPayload.StackStatus {
		stack, err := handler.DataStore.EdgeStack().EdgeStack(stackID)
		if err != nil {
			log.Error().Err(err).Int("stack_id", int(stackID)).Msg("fetch edge stack")

			continue
		}

		// if the stack represents a successful remote update - skip it
		if endpointStatus, ok := stack.Status[endpoint.ID]; ok && endpointStatus.Details.RemoteUpdateSuccess {
			continue
		}

		if stack.EdgeUpdateID != 0 {
			if status.Details.Error {
				err := handler.edgeUpdateService.RemoveActiveSchedule(endpoint.ID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
				if err != nil {
					log.Warn().
						Err(err).
						Msg("Failed to remove active schedule")
				}
			}

			if status.Details.Ok {
				handler.edgeUpdateService.EdgeStackDeployed(endpoint.ID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
			}
		}

		if status.Details.Remove {
			delete(stack.Status, portaineree.EndpointID(status.EndpointID))
		} else {
			stack.Status[portaineree.EndpointID(status.EndpointID)] = portainer.EdgeStackStatus{
				EndpointID: status.EndpointID,
				Details:    status.Details,
				Error:      status.Error,
			}
		}

		err = handler.DataStore.EdgeStack().UpdateEdgeStack(stack.ID, stack)
		if err != nil {
			log.Error().Err(err).Int("stack_id", int(stackID)).Msg("update edge stack")
		}
	}

	// Save edge stack logs
	for _, logs := range snapshotPayload.StackLogs {
		logs.EndpointID = endpoint.ID

		err := handler.DataStore.EdgeStackLog().Update(&logs)
		if err != nil {
			log.Error().
				Err(err).
				Int("stack_id", int(logs.EdgeStackID)).
				Int("endpoint_id", int(logs.EndpointID)).
				Msg("update edge stack")
		}
	}

	// Save edge jobs status
	for jobID, jobPayload := range snapshotPayload.JobsStatus {
		edgeJob, err := handler.DataStore.EdgeJob().EdgeJob(jobID)
		if err != nil {
			log.Error().Err(err).Int("job", int(jobID)).Msg("fetch edge job")

			continue
		}

		err = handler.FileService.StoreEdgeJobTaskLogFileFromBytes(strconv.Itoa(int(jobID)), strconv.Itoa(int(endpoint.ID)), []byte(jobPayload.LogFileContent))
		if err != nil {
			log.Error().Err(err).Int("job", int(jobID)).Msg("save log file")

			continue
		}

		meta := portaineree.EdgeJobEndpointMeta{CollectLogs: false, LogsStatus: portaineree.EdgeJobLogsStatusCollected}
		if _, ok := edgeJob.GroupLogsCollection[endpoint.ID]; ok {
			edgeJob.GroupLogsCollection[endpoint.ID] = meta
		} else {
			edgeJob.Endpoints[endpoint.ID] = meta
		}

		err = handler.DataStore.EdgeJob().UpdateEdgeJob(edgeJob.ID, edgeJob)
		if err != nil {
			log.Error().Err(err).Int("job", int(jobID)).Msg("fetch edge job")
		}
	}

	// Update snapshot
	switch endpoint.Type {
	case portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment:
		err := handler.updateKubernetesSnapshot(endpoint, snapshotPayload, needFullSnapshot)
		if err != nil {
			log.Error().Err(err).Msg("unable to update Kubernetes snapshot")
		}
	case portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment, portaineree.EdgeAgentOnDockerEnvironment:
		err := handler.updateDockerSnapshot(endpoint, snapshotPayload, needFullSnapshot)
		if err != nil {
			log.Error().Err(err).Msg("unable to update Docker snapshot")
		}
	}
}

func (handler *Handler) sendCommandsSince(endpoint *portaineree.Endpoint, commandTimestamp time.Time, timeZone *time.Location) ([]portaineree.EdgeAsyncCommand, error) {
	storedCommands, err := handler.DataStore.EdgeAsyncCommand().EndpointCommands(endpoint.ID)
	if err != nil {
		return nil, err
	}

	var commandsResponse []portaineree.EdgeAsyncCommand
	for _, storedCommand := range storedCommands {
		if storedCommand.Executed {
			continue
		}

		if storedCommand.ScheduledTime != "" {

			pastSchedule, err := shouldScheduleTrigger(storedCommand.ScheduledTime, timeZone)
			if err != nil {
				return nil, err
			}

			if !pastSchedule {
				continue
			}

			storedCommand.Executed = true
			err = handler.DataStore.EdgeAsyncCommand().Update(storedCommand.ID, &storedCommand)
			if err != nil {
				return commandsResponse, err
			}
		} else if !storedCommand.Timestamp.After(commandTimestamp) { // not a scheduled command
			storedCommand.Executed = true
			err := handler.DataStore.EdgeAsyncCommand().Update(storedCommand.ID, &storedCommand)
			if err != nil {
				return commandsResponse, err
			}
			continue
		}

		commandsResponse = append(commandsResponse, storedCommand)
	}

	if len(commandsResponse) > 0 {
		log.Debug().Str("endpoint", endpoint.Name).Time("command_timestamp", commandTimestamp).Msg("sending commands")
	}

	return commandsResponse, nil
}

func (handler *Handler) getEndpoint(endpointID portaineree.EndpointID, edgeID string) (*portaineree.Endpoint, error) {
	if endpointID == 0 {
		var ok bool
		endpointID, ok = handler.DataStore.Endpoint().EndpointIDByEdgeID(edgeID)
		if !ok {
			return nil, errors.New("Unable to retrieve environments from database")
		}

		endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
		if err != nil {
			return nil, errors.WithMessage(err, "Unable to retrieve environments from database")
		}

		return endpoint, nil
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to retrieve the Endpoint from the database")
	}

	if endpoint.EdgeID == "" {
		endpoint.EdgeID = edgeID

		err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
		if err != nil {
			return nil, errors.WithMessage(err, "Unable to update the Endpoint in the database")
		}
	}

	return endpoint, nil
}

func snapshotHash(snapshot []byte) uint32 {
	h := fnv.New32a()
	h.Write(snapshot)

	return h.Sum32()
}
