package edgeasync

import (
	"encoding/base64"
	"fmt"
	"time"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	internaledge "github.com/portainer/portainer-ee/api/internal/edge"
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

type edgeConfigData struct {
	ID         portaineree.EdgeConfigID `json:"id"`
	Name       string                   `json:"name"`
	BaseDir    string                   `json:"baseDir"`
	DirEntries []filesystem.DirEntry    `json:"dirEntries"`
	Prev       *edgeConfigData          `json:"prev,omitempty"`
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
	return service.storeUpdateStackCommand(service.dataStore, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpAdd, scheduledTime, 0)
}

func (service *Service) AddStackCommandTx(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID, scheduledTime string) error {
	return service.storeUpdateStackCommand(tx, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpAdd, scheduledTime, 0)
}

func (service *Service) ReplaceStackCommand(endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID) error {
	return service.storeUpdateStackCommand(service.dataStore, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpReplace, "", 0)
}

func (service *Service) ReplaceStackCommandTx(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID) error {
	return service.storeUpdateStackCommand(tx, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpReplace, "", 0)
}

func (service *Service) ReplaceStackCommandWithVersion(endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID, version int) error {
	return service.storeUpdateStackCommand(service.dataStore, endpoint, edgeStackID, portaineree.EdgeAsyncCommandOpReplace, "", version)
}

func (service *Service) storeUpdateStackCommand(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeStackID portaineree.EdgeStackID, commandOperation portaineree.EdgeAsyncCommandOperation, scheduledTime string, targetVersion int) error {
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

	var (
		projectVersionPath string
		stackFileVersion   int
	)

	rollbackTo := new(int)
	// Check if the requested version is the previous version. If not, use the latest stack file version
	if edgeStack.PreviousDeploymentInfo != nil && targetVersion == edgeStack.PreviousDeploymentInfo.FileVersion {
		projectVersionPath = service.fileService.FormProjectPathByVersion(edgeStack.ProjectPath, edgeStack.PreviousDeploymentInfo.FileVersion, edgeStack.PreviousDeploymentInfo.ConfigHash)
		*rollbackTo = edgeStack.PreviousDeploymentInfo.FileVersion
	} else {
		if targetVersion != 0 && targetVersion != edgeStack.StackFileVersion {
			log.Warn().Msgf("Invalid stack file version %d being requested, fallback to the latest stack file version %d", targetVersion, edgeStack.StackFileVersion)
		}

		// If the requested version is not the previous version, use the latest stack file version
		commitHash := ""
		if edgeStack.GitConfig != nil {
			commitHash = edgeStack.GitConfig.ConfigHash
		}
		projectVersionPath = service.fileService.FormProjectPathByVersion(edgeStack.ProjectPath, edgeStack.StackFileVersion, commitHash)
		stackFileVersion = edgeStack.StackFileVersion
		rollbackTo = nil
	}

	dirEntries, err := filesystem.LoadDir(projectVersionPath)
	if err != nil {
		return httperror.InternalServerError("Unable to load repository", err)
	}

	fileContent, err := filesystem.FilterDirForCompatibility(dirEntries, fileName, endpoint.Agent.Version)
	if err != nil {
		return httperror.InternalServerError("File not found", err)
	}

	if internaledge.IsEdgeStackRelativePathEnabled(edgeStack) {
		if internaledge.IsEdgeStackPerDeviceConfigsEnabled(edgeStack) {
			dirEntries = filesystem.FilterDirForPerDevConfigs(
				dirEntries,
				endpoint.EdgeID,
				edgeStack.PerDeviceConfigsPath,
				edgeStack.PerDeviceConfigsMatchType,
			)
		}
	} else {
		dirEntries = filesystem.FilterDirForEntryFile(dirEntries, fileName)
	}

	registryCredentials := registryutils.GetRegistryCredentialsForEdgeStack(tx, edgeStack, endpoint)

	namespace := ""
	if !edgeStack.UseManifestNamespaces {
		namespace = kubernetes.DefaultNamespace
	}

	envVars := edgeStack.EnvVars
	if envVars == nil {
		envVars = make([]portainer.Pair, 0)
	}

	// If the stack is not for updater, we use stack file version
	version := stackFileVersion
	if edgeStack.EdgeUpdateID > 0 {
		version = edgeStack.Version
	}

	stackStatus := edge.StackPayload{
		DirEntries:          dirEntries,
		EntryFileName:       fileName,
		StackFileContent:    fileContent,
		SupportRelativePath: edgeStack.SupportRelativePath,
		FilesystemPath:      edgeStack.FilesystemPath,
		Name:                edgeStack.Name,
		ID:                  int(edgeStackID),
		Version:             version,
		RollbackTo:          rollbackTo,
		RegistryCredentials: registryCredentials,
		Namespace:           namespace,
		PrePullImage:        edgeStack.PrePullImage,
		RePullImage:         edgeStack.RePullImage,
		RetryDeploy:         edgeStack.RetryDeploy,
		EdgeUpdateID:        edgeStack.EdgeUpdateID,
		EnvVars:             envVars,
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
	return service.RemoveStackCommandTx(service.dataStore, endpointID, edgeStackID)
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
		EnvVars:             edgeStack.EnvVars,
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
func (service *Service) AddConfigCommandTx(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeConfig *portaineree.EdgeConfig, files []filesystem.DirEntry) error {
	return service.storeUpdateConfigCommand(tx, endpointID, edgeConfig, portaineree.EdgeAsyncCommandOpAdd, files, nil)
}

func (service *Service) UpdateConfigCommandTx(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeConfig *portaineree.EdgeConfig, files []filesystem.DirEntry, prevFiles []filesystem.DirEntry) error {
	return service.storeUpdateConfigCommand(tx, endpointID, edgeConfig, portaineree.EdgeAsyncCommandOpReplace, files, prevFiles)
}

func (service *Service) DeleteConfigCommandTx(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeConfig *portaineree.EdgeConfig, files []filesystem.DirEntry) error {
	return service.storeUpdateConfigCommand(tx, endpointID, edgeConfig, portaineree.EdgeAsyncCommandOpRemove, files, nil)
}

func (service *Service) storeUpdateConfigCommand(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, edgeConfig *portaineree.EdgeConfig, commandOperation portaineree.EdgeAsyncCommandOperation, files, prevFiles []filesystem.DirEntry) error {
	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if err != nil || !endpoint.Edge.AsyncMode {
		return err
	}

	payload := &edgeConfigData{
		ID:         edgeConfig.ID,
		Name:       edgeConfig.Name,
		BaseDir:    edgeConfig.BaseDir,
		DirEntries: files,
	}

	if commandOperation == portaineree.EdgeAsyncCommandOpReplace {
		payload.Prev = &edgeConfigData{
			DirEntries: prevFiles,
		}
	}

	asyncCommand := &portaineree.EdgeAsyncCommand{
		Type:       portaineree.EdgeAsyncCommandTypeConfig,
		EndpointID: endpointID,
		Timestamp:  time.Now(),
		Operation:  commandOperation,
		Path:       fmt.Sprintf("/edgeconfig/%d", edgeConfig.ID),
		Value:      payload,
	}

	return tx.EdgeAsyncCommand().Create(asyncCommand)
}
