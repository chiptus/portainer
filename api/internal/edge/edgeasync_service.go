package edge

import (
	"encoding/base64"
	"fmt"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"

	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog/log"
)

type edgeStackData struct {
	ID               portaineree.EdgeStackID
	Version          int
	StackFileContent string
	Name             string
}

type edgeLogData struct {
	EdgeStackID   portaineree.EdgeStackID
	EdgeStackName string
	Tail          int
}

type edgeJobData struct {
	ID                portaineree.EdgeJobID
	CollectLogs       bool
	LogsStatus        portaineree.EdgeJobLogsStatus
	CronExpression    string
	ScriptFileContent string
	Version           int
}

type containerCommandData struct {
	ContainerName          string
	ContainerStartOptions  types.ContainerStartOptions
	ContainerRemoveOptions types.ContainerRemoveOptions
	ContainerOperation     portaineree.EdgeAsyncContainerOperation
}

type imageCommandData struct {
	ImageName          string
	ImageRemoveOptions types.ImageRemoveOptions
	ImageOperation     portaineree.EdgeAsyncImageOperation
}

type volumeCommandData struct {
	VolumeName      string
	ForceRemove     bool
	VolumeOperation portaineree.EdgeAsyncVolumeOperation
}

type Service struct {
	dataStore   dataservices.DataStore
	fileService portainer.FileService
}

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
			log.Error().Msg("Docker is not supported by this stack")

			return nil
		}
	} else if endpointutils.IsKubernetesEndpoint(endpoint) {
		fileName = edgeStack.ManifestPath
		if fileName == "" {
			log.Error().Msg("Kubernetes is not supported by this stack")

			return nil
		}
	}

	stackFileContent, err := service.fileService.GetFileContent(edgeStack.ProjectPath, fileName)
	if err != nil {
		log.Error().Err(err).Msg("unable to retrieve Compose file from disk")

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
	if err != nil || !endpoint.Edge.AsyncMode {
		return err
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

func (service *Service) AddLogCommand(edgeStack *portaineree.EdgeStack, endpointID portaineree.EndpointID, tail int) error {
	cmd := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeLog,
		EndpointID: portaineree.EndpointID(endpointID),
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/edgestack/%d", edgeStack.ID),
		Value: edgeLogData{
			EdgeStackID:   edgeStack.ID,
			EdgeStackName: edgeStack.Name,
			Tail:          tail,
		},
	}

	return service.dataStore.EdgeAsyncCommand().Create(cmd)
}

func (service *Service) AddJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpAdd)
}

func (service *Service) ReplaceJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpReplace)
}

func (service *Service) storeUpdateJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte, commandOperation portaineree.EdgeAsyncCommandOperation) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil || !endpoint.Edge.AsyncMode {
		return err
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
	if err != nil || !endpoint.Edge.AsyncMode {
		return err
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

func (service *Service) StartContainerCommand(endpointID portaineree.EndpointID, name string, startOptions types.ContainerStartOptions) error {
	cmdData := containerCommandData{
		ContainerName:          name,
		ContainerStartOptions:  startOptions,
		ContainerRemoveOptions: types.ContainerRemoveOptions{},
		ContainerOperation:     portaineree.EdgeAsyncContainerOperationStart,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeContainer,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/containers/%s/%s", portaineree.EdgeAsyncContainerOperationStart, name),
		Value:      cmdData,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) RestartContainerCommand(endpointID portaineree.EndpointID, name string) error {
	cmdData := containerCommandData{
		ContainerName:          name,
		ContainerStartOptions:  types.ContainerStartOptions{},
		ContainerRemoveOptions: types.ContainerRemoveOptions{},
		ContainerOperation:     portaineree.EdgeAsyncContainerOperationRestart,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeContainer,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/containers/%s/%s", portaineree.EdgeAsyncContainerOperationRestart, name),
		Value:      cmdData,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) StopContainerCommand(endpointID portaineree.EndpointID, name string) error {
	cmdData := containerCommandData{
		ContainerName:          name,
		ContainerStartOptions:  types.ContainerStartOptions{},
		ContainerRemoveOptions: types.ContainerRemoveOptions{},
		ContainerOperation:     portaineree.EdgeAsyncContainerOperationStop,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeContainer,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/containers/%s/%s", portaineree.EdgeAsyncContainerOperationStop, name),
		Value:      cmdData,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) DeleteContainerCommand(endpointID portaineree.EndpointID, name string, removeOptions types.ContainerRemoveOptions) error {
	cmdData := containerCommandData{
		ContainerName:          name,
		ContainerStartOptions:  types.ContainerStartOptions{},
		ContainerRemoveOptions: removeOptions,
		ContainerOperation:     portaineree.EdgeAsyncContainerOperationDelete,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeContainer,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/containers/%s/%s", portaineree.EdgeAsyncContainerOperationDelete, name),
		Value:      cmdData,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) KillContainerCommand(endpointID portaineree.EndpointID, name string) error {
	cmdData := containerCommandData{
		ContainerName:          name,
		ContainerStartOptions:  types.ContainerStartOptions{},
		ContainerRemoveOptions: types.ContainerRemoveOptions{},
		ContainerOperation:     portaineree.EdgeAsyncContainerOperationKill,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeContainer,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/containers/%s/%s", portaineree.EdgeAsyncContainerOperationKill, name),
		Value:      cmdData,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) DeleteImageCommand(endpointID portaineree.EndpointID, name string, removeOptions types.ImageRemoveOptions) error {
	cmdData := imageCommandData{
		ImageName:          name,
		ImageRemoveOptions: removeOptions,
		ImageOperation:     portaineree.EdgeAsyncImageOperationDelete,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeImage,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/images/%s/%s", portaineree.EdgeAsyncImageOperationDelete, name),
		Value:      cmdData,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}

func (service *Service) DeleteVolumeCommand(endpointID portaineree.EndpointID, name string, forceRemove bool) error {
	cmdData := volumeCommandData{
		VolumeName:      name,
		ForceRemove:     forceRemove,
		VolumeOperation: portaineree.EdgeAsyncVolumeOperationDelete,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeVolume,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/volumes/%s/%s", portaineree.EdgeAsyncVolumeOperationDelete, name),
		Value:      cmdData,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}
