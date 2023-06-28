package edgeasync

import (
	"encoding/base64"
	"fmt"
	"time"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	"github.com/portainer/portainer-ee/api/kubernetes"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/edge"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog/log"
)

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

type normalStackData struct {
	Name             string
	StackFileContent string
	StackOperation   portaineree.EdgeAsyncNormalStackOperation
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

// Deprecated: use AddStackCommandTx instead.
func (service *Service) AddStackCommand(endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID, scheduledTime string) error {
	return service.storeUpdateStackCommand(service.dataStore, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpAdd, scheduledTime)
}

func (service *Service) AddStackCommandTx(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID, scheduledTime string) error {
	return service.storeUpdateStackCommand(tx, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpAdd, scheduledTime)
}

func (service *Service) ReplaceStackCommand(endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID) error {
	return service.storeUpdateStackCommand(service.dataStore, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpReplace, "")
}

func (service *Service) ReplaceStackCommandTx(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID) error {
	return service.storeUpdateStackCommand(tx, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpReplace, "")
}

func (service *Service) storeUpdateStackCommand(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID, commandOperation portaineree.EdgeAsyncCommandOperation, scheduledTime string) error {
	if !endpoint.Edge.AsyncMode {
		return nil
	}

	edgeStack, err := tx.EdgeStack().EdgeStack(edgeStackID)
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

	projectPath := edgeStack.ProjectPath
	if edgeStack.GitConfig == nil {
		projectPath = service.fileService.FormProjectPathByVersion(edgeStack.ProjectPath, edgeStack.Version)
	}
	dirEntries, err := filesystem.LoadDir(projectPath)
	if err != nil {
		return httperror.InternalServerError("Unable to load repository", err)
	}

	if !edgeStack.SupportRelativePath || edgeStack.FilesystemPath == "" {
		dirEntries = filesystem.FilterDirForEntryFile(dirEntries, fileName)
	}

	registryCredentials := registryutils.GetRegistryCredentialsForEdgeStack(tx, edgeStack, endpoint)

	namespace := ""
	if !edgeStack.UseManifestNamespaces {
		namespace = kubernetes.DefaultNamespace
	}

	stackStatus := edge.StackPayload{
		DirEntries:          dirEntries,
		EntryFileName:       fileName,
		SupportRelativePath: edgeStack.SupportRelativePath,
		FilesystemPath:      edgeStack.FilesystemPath,
		Name:                edgeStack.Name,
		ID:                  int(edgeStackID),
		Version:             edgeStack.Version,
		RegistryCredentials: registryCredentials,
		Namespace:           namespace,
		PrePullImage:        edgeStack.PrePullImage,
		RePullImage:         edgeStack.RePullImage,
		RetryDeploy:         edgeStack.RetryDeploy,
		EdgeUpdateID:        edgeStack.EdgeUpdateID,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:          portaineree.EdgeAsyncCommandTypeStack,
		EndpointID:    endpoint.ID,
		Timestamp:     time.Now(),
		Operation:     commandOperation,
		Path:          fmt.Sprintf("/edgestack/%d", edgeStackID),
		Value:         stackStatus,
		ScheduledTime: scheduledTime,
	}

	return tx.EdgeAsyncCommand().Create(asyncCommand)
}

// Deprecated: use RemoveStackCommandTx instead.
func (service *Service) RemoveStackCommand(endpointID portaineree.EndpointID, edgeStackID portaineree.EdgeStackID) error {
	return service.RemoveStackCommandTx(service.dataStore, endpointID, portaineree.EdgeStackID(edgeStackID))
}

func (service *Service) RemoveStackCommandTx(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeStackID portaineree.EdgeStackID) error {
	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if err != nil || !endpoint.Edge.AsyncMode {
		return err
	}

	edgeStack, err := tx.EdgeStack().EdgeStack(edgeStackID)
	if err != nil {
		return err
	}

	stackStatus := edge.StackPayload{
		Name:                edgeStack.Name,
		ID:                  int(edgeStackID),
		Version:             edgeStack.Version,
		EntryFileName:       edgeStack.EntryPoint,
		SupportRelativePath: edgeStack.SupportRelativePath,
		FilesystemPath:      edgeStack.FilesystemPath,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeStack,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpRemove,
		Path:       fmt.Sprintf("/edgestack/%d", edgeStackID),
		Value:      stackStatus,
	}

	return tx.EdgeAsyncCommand().Create(asyncCommand)
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

// Deprecated: use AddJobCommandTx instead
func (service *Service) AddJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(service.dataStore, endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpAdd)
}

func (service *Service) AddJobCommandTx(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(tx, endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpAdd)
}

func (service *Service) ReplaceJobCommand(endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(service.dataStore, endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpReplace)
}

func (service *Service) ReplaceJobCommandTx(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte) error {
	return service.storeUpdateJobCommand(tx, endpointID, edgeJob, fileContent, portaineree.EdgeAsyncCommandOpReplace)
}

func (service *Service) storeUpdateJobCommand(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeJob portaineree.EdgeJob, fileContent []byte, commandOperation portaineree.EdgeAsyncCommandOperation) error {
	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if err != nil || !endpoint.Edge.AsyncMode {
		return err
	}

	collectLogs := edgeJob.Endpoints[endpointID].CollectLogs
	logsStatus := edgeJob.Endpoints[endpointID].LogsStatus

	if v, ok := edgeJob.GroupLogsCollection[endpointID]; ok {
		collectLogs = v.CollectLogs
		logsStatus = v.LogsStatus
	}

	edgeJobData := &edgeJobData{
		ID:                edgeJob.ID,
		CronExpression:    edgeJob.CronExpression,
		CollectLogs:       collectLogs,
		LogsStatus:        logsStatus,
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

	return tx.EdgeAsyncCommand().Create(asyncCommand)
}

// Deprecated: use RemoveJobCommandTx instead
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

func (service *Service) RemoveJobCommandTx(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeJobID portaineree.EdgeJobID) error {
	endpoint, err := tx.Endpoint().Endpoint(endpointID)
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

	return tx.EdgeAsyncCommand().Create(asyncCommand)
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

func (service *Service) RemoveNormalStackCommand(endpointID portaineree.EndpointID, stackID portaineree.StackID) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil || !endpoint.Edge.AsyncMode {
		return err
	}

	stack, err := service.dataStore.Stack().Read(stackID)
	if err != nil {
		return err
	}

	fileContent, err := service.fileService.GetFileContent(stack.ProjectPath, stack.EntryPoint)
	if err != nil {
		return err
	}

	stackStatus := normalStackData{
		Name:             stack.Name,
		StackFileContent: string(fileContent),
		StackOperation:   portaineree.EdgeAsyncNormalStackOperationRemove,
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeNormalStack,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  portaineree.EdgeAsyncCommandOpAdd,
		Path:       fmt.Sprintf("/normalStack/%d", stackID),
		Value:      stackStatus,
	}

	return service.dataStore.EdgeAsyncCommand().Create(asyncCommand)
}
