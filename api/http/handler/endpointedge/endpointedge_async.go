package endpointedge

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge"
	portainer "github.com/portainer/portainer/api"
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
	Docker          *portainer.DockerSnapshot                               `json:"docker,omitempty"`
	DockerPatch     jsonpatch.Patch                                         `json:"dockerPatch,omitempty"`
	Kubernetes      *portaineree.KubernetesSnapshot                         `json:"kubernetes,omitempty"`
	KubernetesPatch jsonpatch.Patch                                         `json:"kubernetesPatch,omitempty"`
	StackStatus     map[portaineree.EdgeStackID]portaineree.EdgeStackStatus `json:"stackStatus,omitempty"`
	JobsStatus      map[portaineree.EdgeJobID]portaineree.EdgeJobStatus     `json:"jobsStatus:,omitempty"`
}

func (payload *EdgeAsyncRequest) Validate(r *http.Request) error {
	return nil
}

type EdgeAsyncResponse struct {
	EndpointID portaineree.EndpointID `json:"endpointID"`

	PingInterval     time.Duration `json:"pingInterval"`
	SnapshotInterval time.Duration `json:"snapshotInterval"`
	CommandInterval  time.Duration `json:"commandInterval"`

	Commands []portaineree.EdgeAsyncCommand `json:"commands"`
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
		logrus.WithField("PortainerAgentEdgeIDHeader", edgeID).Debug("missing agent edge id")
		return httperror.BadRequest("missing Edge identifier", errors.New("missing Edge identifier"))
	}

	var payload EdgeAsyncRequest
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		logrus.WithError(err).WithField("payload", r).Debug("decode payload")
		return httperror.BadRequest("Invalid request payload", err)
	}

	endpoint, err := handler.getEndpoint(payload.EndpointID, edgeID)
	if err != nil {
		return httperror.InternalServerError("Endpoint with edge id or endpoint id is missing", err)
	}

	if endpoint == nil {
		logrus.WithField("PortainerAgentEdgeIDHeader", edgeID).Debug("edge id not found in existing endpoints")
		agentPlatform, agentPlatformErr := parseAgentPlatform(r)
		if agentPlatformErr != nil {
			return httperror.BadRequest("agent platform header is not valid", err)
		}

		validateCertsErr := handler.requestBouncer.AuthorizedClientTLSConn(r)
		if validateCertsErr != nil {
			return httperror.Forbidden("Permission denied to access environment", err)
		}

		newEndpoint, createEndpointErr := handler.createAsyncEdgeAgentEndpoint(r, edgeID, agentPlatform)
		if createEndpointErr != nil {
			return createEndpointErr
		}
		endpoint = newEndpoint

		asyncResponse := EdgeAsyncResponse{
			EndpointID: endpoint.ID,
		}

		return response.JSON(w, asyncResponse)
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	if r.Header.Get("Content-Encoding") == "gzip" {
		gzr, err := gzip.NewReader(r.Body)
		if err != nil {
			return httperror.BadRequest("Invalid request payload", err)
		}

		r.Body = gzr
	}

	if payload.Snapshot != nil {
		handler.saveSnapshot(endpoint, payload.Snapshot)
	}

	endpoint.LastCheckInDate = time.Now().Unix()
	endpoint.Status = portaineree.EndpointStatusUp
	endpoint.Edge.AsyncMode = true
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
	}

	if payload.CommandTimestamp != nil {
		commands, err := handler.sendCommandsSince(endpoint, *payload.CommandTimestamp)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve commands", err)
		}

		asyncResponse.Commands = commands
	}

	return response.JSON(w, asyncResponse)
}

func (handler *Handler) createAsyncEdgeAgentEndpoint(req *http.Request, edgeID string, endpointType portaineree.EndpointType) (*portaineree.Endpoint, *httperror.HandlerError) {
	endpointID := handler.DataStore.Endpoint().GetNextIdentifier()

	requestURL := fmt.Sprintf("https://%s", req.Host)
	portainerURL, err := url.Parse(requestURL)
	if err != nil {
		return nil, httperror.BadRequest("Invalid environment URL", err)
	}
	portainerHost, _, err := net.SplitHostPort(portainerURL.Host)
	if err != nil {
		portainerHost = portainerURL.Host
	}

	edgeKey := handler.ReverseTunnelService.GenerateEdgeKey(requestURL, portainerHost, endpointID)

	endpoint := &portaineree.Endpoint{
		ID:     portaineree.EndpointID(endpointID),
		EdgeID: edgeID,
		Name:   edgeID,
		URL:    requestURL,
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

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve the settings", err)
	}

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

func (handler *Handler) updateDockerSnapshot(endpoint *portaineree.Endpoint, snapshotPayload *snapshot) error {
	if len(snapshotPayload.DockerPatch) > 0 {
		if len(endpoint.Snapshots) == 0 {
			return errors.New("received a Docker snapshot patch but there was no previous snapshot")
		}

		lastSnapshot := endpoint.Snapshots[len(endpoint.Snapshots)-1]

		lastSnapshotJSON, err := json.Marshal(lastSnapshot)
		if err != nil {
			return fmt.Errorf("could not marshal the last Docker snapshot: %s", err)
		}

		newSnapshotJSON, err := snapshotPayload.DockerPatch.Apply(lastSnapshotJSON)
		if err != nil {
			return fmt.Errorf("could not apply the patch to the Docker snapshot: %s", err)
		}

		var newSnapshot portainer.DockerSnapshot
		err = json.Unmarshal(newSnapshotJSON, &newSnapshot)
		if err != nil {
			return fmt.Errorf("could not unmarshal the patched Docker snapshot: %s", err)
		}

		logrus.Debug("setting the patched Docker snapshot")
		endpoint.Snapshots = []portainer.DockerSnapshot{newSnapshot}
	} else if snapshotPayload.Docker != nil {
		logrus.Debug("setting the full Docker snapshot")
		endpoint.Snapshots = []portainer.DockerSnapshot{*snapshotPayload.Docker}
	}

	return nil
}

func (handler *Handler) updateKubernetesSnapshot(endpoint *portaineree.Endpoint, snapshotPayload *snapshot) error {
	if len(snapshotPayload.KubernetesPatch) > 0 {
		if len(endpoint.Kubernetes.Snapshots) == 0 {
			return errors.New("received a Kubernetes snapshot patch but there was no previous snapshot")
		}

		lastSnapshot := endpoint.Kubernetes.Snapshots[len(endpoint.Kubernetes.Snapshots)-1]

		lastSnapshotJSON, err := json.Marshal(lastSnapshot)
		if err != nil {
			return fmt.Errorf("could not marshal the last Kubernetes snapshot: %s", err)
		}

		newSnapshotJSON, err := snapshotPayload.KubernetesPatch.Apply(lastSnapshotJSON)
		if err != nil {
			return fmt.Errorf("could not apply the patch to the Kubernetes snapshot: %s", err)
		}

		var newSnapshot portaineree.KubernetesSnapshot
		err = json.Unmarshal(newSnapshotJSON, &newSnapshot)
		if err != nil {
			return fmt.Errorf("could not unmarshal the patched Kubernetes snapshot: %s", err)
		}

		logrus.Debug("setting the patched Kubernetes snapshot")
		endpoint.Kubernetes.Snapshots = []portaineree.KubernetesSnapshot{newSnapshot}
	} else if snapshotPayload.Kubernetes != nil {
		logrus.Debug("setting the full Kubernetes snapshot")
		endpoint.Kubernetes.Snapshots = []portaineree.KubernetesSnapshot{*snapshotPayload.Kubernetes}
	}

	return nil
}

func (handler *Handler) saveSnapshot(endpoint *portaineree.Endpoint, snapshotPayload *snapshot) {
	// Save edge stacks status
	for stackID, status := range snapshotPayload.StackStatus {
		stack, err := handler.DataStore.EdgeStack().EdgeStack(stackID)
		if err != nil {
			logrus.WithError(err).WithField("stack", stackID).Error("fetch edge stack")
			continue
		}

		if status.Type == portaineree.EdgeStackStatusRemove {
			delete(stack.Status, status.EndpointID)
		} else {
			stack.Status[status.EndpointID] = portaineree.EdgeStackStatus{
				EndpointID: status.EndpointID,
				Type:       status.Type,
				Error:      status.Error,
			}
		}

		err = handler.DataStore.EdgeStack().UpdateEdgeStack(stack.ID, stack)
		if err != nil {
			logrus.WithError(err).WithField("stack", stackID).Error("update edge stack")
		}
	}

	// Save edge jobs status
	for jobID, jobPayload := range snapshotPayload.JobsStatus {
		edgeJob, err := handler.DataStore.EdgeJob().EdgeJob(jobID)
		if err != nil {
			logrus.WithError(err).WithField("job", jobID).Error("fetch edge job")
			continue
		}

		err = handler.FileService.StoreEdgeJobTaskLogFileFromBytes(strconv.Itoa(int(jobID)), strconv.Itoa(int(endpoint.ID)), []byte(jobPayload.LogFileContent))
		if err != nil {
			logrus.WithError(err).WithField("job", jobID).Error("save log file")
			continue
		}

		meta := edgeJob.Endpoints[endpoint.ID]
		meta.CollectLogs = false
		meta.LogsStatus = portaineree.EdgeJobLogsStatusCollected
		edgeJob.Endpoints[endpoint.ID] = meta

		err = handler.DataStore.EdgeJob().UpdateEdgeJob(edgeJob.ID, edgeJob)
		if err != nil {
			logrus.WithError(err).WithField("job", jobID).Error("fetch edge job")
		}
	}

	// Update snapshot
	switch endpoint.Type {
	case portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment:
		err := handler.updateKubernetesSnapshot(endpoint, snapshotPayload)
		if err != nil {
			logrus.Error(err)
		}
	case portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment, portaineree.EdgeAgentOnDockerEnvironment:
		err := handler.updateDockerSnapshot(endpoint, snapshotPayload)
		if err != nil {
			logrus.Error(err)
		}
	}
}

func (handler *Handler) sendCommandsSince(endpoint *portaineree.Endpoint, commandTimestamp time.Time) ([]portaineree.EdgeAsyncCommand, error) {
	storedCommands, err := handler.DataStore.EdgeAsyncCommand().EndpointCommands(endpoint.ID)
	if err != nil {
		return nil, err
	}

	var commandsResponse []portaineree.EdgeAsyncCommand
	for _, storedCommand := range storedCommands {
		if storedCommand.Executed {
			continue
		}

		if commandTimestamp.After(storedCommand.Timestamp) {
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
		logrus.WithField("endpoint", endpoint.Name).WithField("from command", commandTimestamp).Debug("Sending commands")
	}

	return commandsResponse, nil
}

func (handler *Handler) getEndpoint(endpointID portaineree.EndpointID, edgeID string) (*portaineree.Endpoint, error) {
	if endpointID == 0 {
		endpoints, err := handler.DataStore.Endpoint().Endpoints()
		if err != nil {
			return nil, errors.WithMessage(err, "Unable to retrieve environments from database")
		}

		return edge.EdgeEndpoint(endpoints, edgeID), nil
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
