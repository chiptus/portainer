package edge

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"
)

type edgeStackData struct {
	ID               portaineree.EdgeStackID
	Version          int
	StackFileContent string
	Name             string
}

type edgeJobData struct {
	ID                portaineree.EdgeJobID
	CollectLogs       bool
	LogsStatus        portaineree.EdgeJobLogsStatus
	CronExpression    string
	ScriptFileContent string
	Version           int
}

type (
	Service struct {
		dataStore   dataservices.DataStore
		fileService portainer.FileService
	}
)

func NewService(dataStore dataservices.DataStore, fileService portainer.FileService) *Service {
	return &Service{
		dataStore:   dataStore,
		fileService: fileService,
	}
}

func (service *Service) AddStackCommand(endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID) error {
	return service.storeUpdateStackCommand(endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpAdd)
}

func (service *Service) ReplaceStackCommand(endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID) error {
	return service.storeUpdateStackCommand(endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpReplace)
}

func (service *Service) storeUpdateStackCommand(endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID, commandOperation portaineree.EdgeAsyncCommandOperation) error {
	if !endpoint.Edge.AsyncMode {
		return nil
	}

	edgeStack, err := service.dataStore.EdgeStack().EdgeStack(edgeStackID)
	if err != nil {
		return err
	}

	fileName := edgeStack.EntryPoint
	if endpointutils.IsDockerEndpoint(endpoint) {
		if fileName == "" {
			logrus.Error("Docker is not supported by this stack")
			return nil
		}
	}
	if endpointutils.IsKubernetesEndpoint(endpoint) {
		fileName = edgeStack.ManifestPath
		if fileName == "" {
			logrus.Error("Kubernetes is not supported by this stack")
			return nil
		}
	}

	stackFileContent, err := service.fileService.GetFileContent(edgeStack.ProjectPath, fileName)
	if err != nil {
		logrus.WithError(err).Error("Unable to retrieve Compose file from disk")
		return err
	}

	stackStatus := edgeStackData{
		StackFileContent: string(stackFileContent),
		Name:             edgeStack.Name,
		ID:               edgeStackID,
		Version:          edgeStack.Version,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeStack,
		EndpointID: endpoint.ID,
		Timestamp:  time.Now(),
		Operation:  commandOperation,
		Path:       fmt.Sprintf("/edgestack/%d", edgeStackID),
		Value:      stackStatus,
	}
	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) RemoveStackCommand(endpointID portaineree.EndpointID, edgeStackID portaineree.EdgeStackID) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return err
	}

	if !endpoint.Edge.AsyncMode {
		return nil
	}

	edgeStack, err := service.dataStore.EdgeStack().EdgeStack(edgeStackID)
	if err != nil {
		return err
	}

	stackStatus := edgeStackData{
		Name:    edgeStack.Name,
		ID:      edgeStackID,
		Version: edgeStack.Version,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeStack,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpRemove,
		Path:       fmt.Sprintf("/edgestack/%d", edgeStackID),
		Value:      stackStatus,
	}
	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) AddJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpAdd)
}

func (service *Service) ReplaceJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpReplace)
}

func (service *Service) storeUpdateJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte, commandOperation portaineree.EdgeAsyncCommandOperation) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return err
	}

	if !endpoint.Edge.AsyncMode {
		return nil
	}

	edgeJobData := &edgeJobData{
		ID:                edgeJob.ID,
		CronExpression:    edgeJob.CronExpression,
		CollectLogs:       edgeJob.Endpoints[endpointID].CollectLogs,
		LogsStatus:        edgeJob.Endpoints[endpointID].LogsStatus,
		Version:           edgeJob.Version,
		ScriptFileContent: base64.RawStdEncoding.EncodeToString(fileContent),
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeJob,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  commandOperation,
		Path:       fmt.Sprintf("/edgejob/%d", edgeJob.ID),
		Value:      edgeJobData,
	}
	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) RemoveJobCommand(endpointID portaineree.EndpointID, edgeJobID portaineree.EdgeJobID) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return err
	}

	if !endpoint.Edge.AsyncMode {
		return nil
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeJob,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpRemove,
		Path:       fmt.Sprintf("/edgejob/%d", edgeJobID),
	}
	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}
