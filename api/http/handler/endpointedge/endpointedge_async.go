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
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/pkg/featureflags"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const edgeIntervalUseDefault = -1

type MetaFields struct {
	EdgeGroupsIDs      []portaineree.EdgeGroupID   `json:"edgeGroupsIds"`
	TagsIDs            []portaineree.TagID         `json:"tagsIds"`
	EnvironmentGroupID portaineree.EndpointGroupID `json:"environmentGroupId"`
}

type EdgeAsyncRequest struct {
	CommandTimestamp *time.Time             `json:"commandTimestamp"`
	Snapshot         *snapshot              `json:"snapshot"`
	EndpointID       portaineree.EndpointID `json:"endpointId"`
	MetaFields       *MetaFields            `json:"metaFields"`
}

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

type EdgeAsyncResponse struct {
	EndpointID portaineree.EndpointID `json:"endpointID"`

	PingInterval     time.Duration `json:"pingInterval"`
	SnapshotInterval time.Duration `json:"snapshotInterval"`
	CommandInterval  time.Duration `json:"commandInterval"`

	Commands         []portaineree.EdgeAsyncCommand `json:"commands"`
	NeedFullSnapshot bool                           `json:"needFullSnapshot"`
}

var errHashMismatch = errors.New("the snapshot hash does not match against the stored one")

func (payload *EdgeAsyncRequest) Validate(r *http.Request) error {
	return nil
}

// @id endpointEdgeAsync
// @summary Get environment(endpoint) status
// @description Environment(Endpoint) for edge agent to check status of environment(endpoint)
// @description **Access policy**: restricted only to Edge environments(endpoints)
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @success 200 {object} EdgeAsyncResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access environment(endpoint)"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/edge/async [post]
func (handler *Handler) endpointEdgeAsync(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeID := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)
	if edgeID == "" {
		log.Debug().Str("PortainerAgentEdgeIDHeader", edgeID).Msg("missing agent edge id")

		return httperror.BadRequest("missing Edge identifier", errors.New("missing Edge identifier"))
	}

	payload, err := parseBodyPayload(r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var asyncResponse *EdgeAsyncResponse
	if featureflags.IsEnabled(portainer.FeatureNoTx) {
		asyncResponse, err = handler.getStatusAsync(handler.DataStore, edgeID, payload, r)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			asyncResponse, err = handler.getStatusAsync(tx, edgeID, payload, r)
			return err
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Uexpected error", err)
	}

	return response.JSON(w, asyncResponse)
}

func (handler *Handler) getStatusAsync(tx dataservices.DataStoreTx, edgeID string, payload EdgeAsyncRequest, r *http.Request) (*EdgeAsyncResponse, error) {
	endpoint, err := handler.getEndpoint(tx, payload.EndpointID, edgeID)
	if err != nil {
		return nil, httperror.InternalServerError("Endpoint with edge ID or endpoint ID is missing", err)
	}

	version := r.Header.Get(portaineree.PortainerAgentHeader)

	timeZone := r.Header.Get(portaineree.PortainerAgentTimeZoneHeader)

	agentPlatform, agentPlatformErr := parseAgentPlatform(r)
	if agentPlatformErr != nil {
		return nil, httperror.BadRequest("agent platform header is not valid", agentPlatformErr)
	}

	err = handler.requestBouncer.AuthorizedClientTLSConn(r)
	if err != nil {
		return nil, httperror.Forbidden("Permission denied to access environment", err)
	}

	if endpoint == nil {
		log.Debug().Str("PortainerAgentEdgeIDHeader", edgeID).Msg("edge id not found in existing endpoints")

		newEndpoint, createEndpointErr := handler.createAsyncEdgeAgentEndpoint(tx, r, edgeID, agentPlatform, version, timeZone, payload.MetaFields)
		if createEndpointErr != nil {
			return nil, createEndpointErr
		}

		asyncResponse := EdgeAsyncResponse{
			EndpointID: newEndpoint.ID,
		}

		return &asyncResponse, nil
	}

	endpoint.Type = agentPlatform

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return nil, httperror.Forbidden("Permission denied to access environment", err)
	}

	updateID, err := parseEdgeUpdateID(r)
	if err != nil {
		return nil, httperror.BadRequest("Unable to parse edge update id", err)
	}

	// check endpoint version, if it has the same version as the active schedule, then we can mark the edge stack as successfully deployed
	activeUpdateSchedule := handler.edgeUpdateService.ActiveSchedule(endpoint.ID)
	if activeUpdateSchedule != nil && activeUpdateSchedule.ScheduleID == updateID {
		err := handler.handleSuccessfulUpdate(tx, activeUpdateSchedule)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to handle successful update", err)
		}

		err = handler.EdgeService.RemoveStackCommandTx(tx, endpoint.ID, activeUpdateSchedule.EdgeStackID)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to replace stack command", err)
		}
	}

	var needFullSnapshot bool
	if payload.Snapshot != nil {
		handler.saveSnapshot(tx, endpoint, payload.Snapshot, &needFullSnapshot)
	}

	if timeZone != "" {
		endpoint.LocalTimeZone = timeZone
	}
	endpoint.LastCheckInDate = time.Now().Unix()
	endpoint.Status = portaineree.EndpointStatusUp
	endpoint.Edge.AsyncMode = true
	endpoint.Agent.Version = version

	err = tx.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to Unable to persist environment changes inside the database", err)
	}

	settings, err := tx.Settings().Settings()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve settings from the database", err)
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
			return nil, httperror.InternalServerError("Unable to parse location", err)
		}

		commands, err := handler.sendCommandsSince(tx, endpoint, *payload.CommandTimestamp, location)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve commands", err)
		}

		asyncResponse.Commands = commands
	}

	return &asyncResponse, nil
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

func (handler *Handler) createAsyncEdgeAgentEndpoint(tx dataservices.DataStoreTx, req *http.Request, edgeID string, endpointType portaineree.EndpointType, version string, timeZone string, metaFields *MetaFields) (*portaineree.Endpoint, *httperror.HandlerError) {
	settings, err := tx.Settings().Settings()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve the settings", err)
	}

	if settings.EdgePortainerURL == "" {
		return nil, httperror.InternalServerError("Portainer API server URL is not set in Edge Compute settings", errors.New("Portainer API server URL is not set in Edge Compute settings"))
	}

	endpointID := tx.Endpoint().GetNextIdentifier()

	edgeKey := handler.ReverseTunnelService.GenerateEdgeKey(settings.EdgePortainerURL, settings.Edge.TunnelServerAddress, endpointID)

	endpoint := &portaineree.Endpoint{
		ID:     portaineree.EndpointID(endpointID),
		EdgeID: edgeID,
		Name:   edgeID,
		URL:    settings.EdgePortainerURL,
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

	var edgeGroupsIDs []portaineree.EdgeGroupID
	if metaFields != nil {
		// validate the environment group
		_, err = tx.EndpointGroup().Read(metaFields.EnvironmentGroupID)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to retrieve the environment group from the database")
			metaFields.EnvironmentGroupID = 1
		}
		endpoint.GroupID = metaFields.EnvironmentGroupID

		// validate tags
		tagsIDs := []portaineree.TagID{}
		for _, tagID := range metaFields.TagsIDs {
			_, err = tx.Tag().Read(tagID)
			if err != nil {
				log.Warn().Err(err).Msg("Unable to retrieve the tag from the database")
				continue
			}

			tagsIDs = append(tagsIDs, tagID)
		}
		endpoint.TagIDs = tagsIDs

		// validate edge groups
		for _, edgeGroupID := range metaFields.EdgeGroupsIDs {
			_, err = tx.EdgeGroup().Read(edgeGroupID)
			if err != nil {
				log.Warn().Err(err).Msg("Unable to retrieve the edge group from the database")
				continue
			}

			edgeGroupsIDs = append(edgeGroupsIDs, edgeGroupID)
		}
	}

	endpoint.Agent.Version = version

	endpoint.Edge.AsyncMode = true
	endpoint.Edge.PingInterval = settings.Edge.PingInterval
	endpoint.Edge.SnapshotInterval = settings.Edge.SnapshotInterval
	endpoint.Edge.CommandInterval = settings.Edge.CommandInterval
	endpoint.UserTrusted = settings.TrustOnFirstConnect

	err = tx.Endpoint().Create(endpoint)
	if err != nil {
		return nil, httperror.InternalServerError("An error occurred while trying to create the environment", err)
	}

	err = edge.AddEnvironmentToEdgeGroups(tx, endpoint, edgeGroupsIDs)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to add environment to edge groups", err)
	}

	for _, tagID := range endpoint.TagIDs {
		if featureflags.IsEnabled(portainer.FeatureNoTx) {
			err = tx.Tag().UpdateTagFunc(tagID, func(tag *portaineree.Tag) {
				tag.Endpoints[endpoint.ID] = true
			})
			if err != nil {
				return endpoint, httperror.InternalServerError("Unable to associate the environment to the specified tag", err)
			}

			continue
		}

		tag, err := tx.Tag().Read(tagID)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve tag from the database", err)
		}

		tag.Endpoints[endpoint.ID] = true

		err = tx.Tag().Update(tag.ID, tag)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to associate the environment to the specified tag", err)
		}
	}

	return endpoint, nil
}

func (handler *Handler) updateDockerSnapshot(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, snapshotPayload *snapshot, needFullSnapshot *bool) error {
	snapshot := &portaineree.Snapshot{}

	if len(snapshotPayload.DockerPatch) > 0 {
		var err error
		snapshot, err = tx.Snapshot().Read(endpoint.ID)
		if err != nil || snapshot.Docker == nil {
			*needFullSnapshot = true
			return errors.New("received a Docker snapshot patch but there was no previous snapshot")
		}

		lastSnapshot := snapshot.Docker
		lastSnapshotJSON, err := json.Marshal(lastSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not marshal the last Docker snapshot: %w", err)
		}

		if snapshotPayload.DockerHash == nil {
			*needFullSnapshot = true
			return fmt.Errorf("the snapshot hash is missing")
		}

		h := snapshotHash(lastSnapshotJSON)
		if *snapshotPayload.DockerHash != h {
			*needFullSnapshot = true
			log.Debug().Uint32("expected", h).Uint32("got", *snapshotPayload.DockerHash).Msg("hash mismatch")

			return errHashMismatch
		}

		newSnapshotJSON, err := snapshotPayload.DockerPatch.Apply(lastSnapshotJSON)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not apply the patch to the Docker snapshot: %w", err)
		}

		var newSnapshot portainer.DockerSnapshot
		err = json.Unmarshal(newSnapshotJSON, &newSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not unmarshal the patched Docker snapshot: %w", err)
		}

		log.Debug().Msg("setting the patched Docker snapshot")
		snapshot.Docker = &newSnapshot
	} else if snapshotPayload.Docker != nil {
		log.Debug().Msg("setting the full Docker snapshot")
		snapshot.Docker = snapshotPayload.Docker
	}

	snapshot.EndpointID = endpoint.ID
	err := tx.Snapshot().Update(endpoint.ID, snapshot)
	if err != nil {
		return errors.New("snapshot could not be updated")
	}

	return nil
}

func (handler *Handler) updateKubernetesSnapshot(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, snapshotPayload *snapshot, needFullSnapshot *bool) error {
	snapshot := &portaineree.Snapshot{}

	if len(snapshotPayload.KubernetesPatch) > 0 {
		var err error
		snapshot, err = tx.Snapshot().Read(endpoint.ID)
		if err != nil || snapshot.Kubernetes == nil {
			*needFullSnapshot = true
			return errors.New("received a Kubernetes snapshot patch but there was no previous snapshot")
		}

		lastSnapshot := snapshot.Kubernetes
		lastSnapshotJSON, err := json.Marshal(lastSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not marshal the last Kubernetes snapshot: %w", err)
		}

		if snapshotPayload.KubernetesHash == nil {
			*needFullSnapshot = true
			return fmt.Errorf("the snapshot hash is missing")
		}

		h := snapshotHash(lastSnapshotJSON)
		if *snapshotPayload.KubernetesHash != h {
			*needFullSnapshot = true
			log.Debug().Uint32("expected", h).Uint32("got", *snapshotPayload.KubernetesHash).Msg("hash mismatch")

			return errHashMismatch
		}

		newSnapshotJSON, err := snapshotPayload.KubernetesPatch.Apply(lastSnapshotJSON)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not apply the patch to the Kubernetes snapshot: %w", err)
		}

		var newSnapshot portaineree.KubernetesSnapshot
		err = json.Unmarshal(newSnapshotJSON, &newSnapshot)
		if err != nil {
			*needFullSnapshot = true
			return fmt.Errorf("could not unmarshal the patched Kubernetes snapshot: %w", err)
		}

		log.Debug().Msg("setting the patched Kubernetes snapshot")
		snapshot.Kubernetes = &newSnapshot
	} else if snapshotPayload.Kubernetes != nil {
		log.Debug().Msg("setting the full Kubernetes snapshot")
		snapshot.Kubernetes = snapshotPayload.Kubernetes
	}

	snapshot.EndpointID = endpoint.ID
	err := tx.Snapshot().Update(endpoint.ID, snapshot)
	if err != nil {
		return errors.New("snapshot could not be updated")
	}

	return nil
}

func (handler *Handler) saveSnapshot(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, snapshotPayload *snapshot, needFullSnapshot *bool) {
	// Save edge stacks status
	for stackID, status := range snapshotPayload.StackStatus {
		stack, err := tx.EdgeStack().EdgeStack(stackID)
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
				err = handler.edgeUpdateService.RemoveActiveSchedule(endpoint.ID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
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

		var deploymentInfo portainer.StackDeploymentInfo
		if status.Details.Ok {
			deploymentInfo.FileVersion = stack.StackFileVersion
			if stack.GitConfig != nil {
				deploymentInfo.ConfigHash = stack.GitConfig.ConfigHash
			}
		}

		if status.Details.Remove {
			delete(stack.Status, portaineree.EndpointID(status.EndpointID))
		} else {
			stack.Status[portaineree.EndpointID(status.EndpointID)] = portainer.EdgeStackStatus{
				EndpointID:     status.EndpointID,
				Details:        status.Details,
				Error:          status.Error,
				DeploymentInfo: deploymentInfo,
			}
		}

		err = tx.EdgeStack().UpdateEdgeStack(stack.ID, stack)
		if err != nil {
			log.Error().Err(err).Int("stack_id", int(stackID)).Msg("update edge stack")
		}
	}

	// Save edge stack logs
	for _, logs := range snapshotPayload.StackLogs {
		logs.EndpointID = endpoint.ID

		err := tx.EdgeStackLog().Update(&logs)
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
		edgeJob, err := tx.EdgeJob().Read(jobID)
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

		err = tx.EdgeJob().Update(edgeJob.ID, edgeJob)
		if err != nil {
			log.Error().Err(err).Int("job", int(jobID)).Msg("fetch edge job")
		}
	}

	// Update snapshot
	switch endpoint.Type {
	case portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment:
		err := handler.updateKubernetesSnapshot(tx, endpoint, snapshotPayload, needFullSnapshot)
		if err != nil && !errors.Is(err, errHashMismatch) {
			log.Error().Err(err).Msg("unable to update Kubernetes snapshot")
		}
	case portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment, portaineree.EdgeAgentOnDockerEnvironment:
		err := handler.updateDockerSnapshot(tx, endpoint, snapshotPayload, needFullSnapshot)
		if err != nil && !errors.Is(err, errHashMismatch) {
			log.Error().Err(err).Msg("unable to update Docker snapshot")
		}
	}
}

func (handler *Handler) sendCommandsSince(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, commandTimestamp time.Time, timeZone *time.Location) ([]portaineree.EdgeAsyncCommand, error) {
	storedCommands, err := tx.EdgeAsyncCommand().EndpointCommands(endpoint.ID)
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
			err = tx.EdgeAsyncCommand().Update(storedCommand.ID, &storedCommand)
			if err != nil {
				return commandsResponse, err
			}
		} else if !storedCommand.Timestamp.After(commandTimestamp) { // not a scheduled command
			storedCommand.Executed = true
			err := tx.EdgeAsyncCommand().Update(storedCommand.ID, &storedCommand)
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

func (handler *Handler) getEndpoint(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeID string) (*portaineree.Endpoint, error) {
	if endpointID == 0 {
		var ok bool
		endpointID, ok = handler.DataStore.Endpoint().EndpointIDByEdgeID(edgeID)
		if !ok {
			return nil, nil
		}

		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return nil, errors.WithMessage(err, "Unable to retrieve environment")
		}

		return endpoint, nil
	}

	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to retrieve the Endpoint from the database")
	}

	if endpoint.EdgeID == "" {
		endpoint.EdgeID = edgeID

		err = tx.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
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
