package cloud

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	clouderrors "github.com/portainer/portainer-ee/api/cloud/errors"
	"github.com/portainer/portainer-ee/api/cloud/gke"
	mk8s "github.com/portainer/portainer-ee/api/cloud/microk8s"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/kubernetes"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/portainer/portainer/api/filesystem"
	log "github.com/rs/zerolog/log"
)

type ProvisioningState int

//go:generate stringer -type=ProvisioningState -trimprefix=ProvisioningState
const (
	ProvisioningStatePending ProvisioningState = iota
	ProvisioningStateWaitingForCluster
	ProvisioningStateDeployingCustomTemplate
	ProvisioningStateAgentSetup
	ProvisioningStateWaitingForAgent
	ProvisioningStateUpdatingEnvironment
	ProvisioningStateDone
)

const (
	stateWaitTime            = 30 * time.Second
	maxRequestFailures       = 240
	maxRequestFailuresImport = 30
)

type (
	CloudManagementService struct {
		dataStore            dataservices.DataStore
		fileService          portaineree.FileService
		shutdownCtx          context.Context
		requests             chan portaineree.CloudManagementRequest
		result               chan *cloudPrevisioningResult
		snapshotService      portaineree.SnapshotService
		authorizationService *authorization.Service
		clientFactory        *kubecli.ClientFactory
		kubernetesDeployer   portaineree.KubernetesDeployer
	}

	cloudPrevisioningResult struct {
		endpointID portaineree.EndpointID
		state      int
		errSummary string
		err        error
		taskID     portaineree.CloudProvisioningTaskID
		provider   string
	}

	CloudProvisioningRequest struct {
		Credentials                                                 *models.CloudCredential
		Region, ClusterName, NodeSize, NetworkID, KubernetesVersion string
		NodeCount                                                   int
	}
)

func NewCloudClusterManagementService(dataStore dataservices.DataStore, fileService portaineree.FileService, clientFactory *kubecli.ClientFactory, snapshotService portaineree.SnapshotService, authorizationService *authorization.Service, shutdownCtx context.Context, kubernetesDeployer portaineree.KubernetesDeployer) *CloudManagementService {
	requests := make(chan portaineree.CloudManagementRequest, 10)
	result := make(chan *cloudPrevisioningResult, 10)

	return &CloudManagementService{
		dataStore:            dataStore,
		fileService:          fileService,
		shutdownCtx:          shutdownCtx,
		requests:             requests,
		result:               result,
		snapshotService:      snapshotService,
		authorizationService: authorizationService,
		clientFactory:        clientFactory,
		kubernetesDeployer:   kubernetesDeployer,
	}
}

func (service *CloudManagementService) Start() {
	log.Info().Msg("starting cloud cluster setup service")

	service.restoreTasks()

	go func() {
		for {
			select {
			case request := <-service.requests:
				go service.processManagementRequest(request)

			case result := <-service.result:
				service.processResult(result)

			case <-service.shutdownCtx.Done():
				log.Debug().Msg("shutting down KaaS setup queue")
				return
			}
		}
	}()
}

// SubmitRequest takes a CloudProvisioningRequest and adds it to the queue to be processed concurrently
func (service *CloudManagementService) SubmitRequest(r portaineree.CloudManagementRequest) {
	service.requests <- r
}

// createClusterSetupTask transforms a provisioning request into a task and adds it to the db
func (service *CloudManagementService) createClusterSetupTask(request *portaineree.CloudProvisioningRequest, clusterID, resourceGroup string) (portaineree.CloudProvisioningTask, error) {
	task := portaineree.CloudProvisioningTask{
		Provider:              request.Provider,
		ClusterID:             clusterID,
		EndpointID:            request.EndpointID,
		Region:                request.Region,
		State:                 request.StartingState,
		CreatedAt:             time.Now(),
		ResourceGroup:         resourceGroup,
		CustomTemplateID:      request.CustomTemplateID,
		CustomTemplateContent: request.CustomTemplateContent,
		MasterNodes:           request.MasterNodes,
		WorkerNodes:           request.WorkerNodes,
		CreatedByUserID:       request.CreatedByUserID,
	}

	return task, service.dataStore.CloudProvisioning().Create(&task)
}

// restoreProvisioningTasks looks up provisioning tasks and retores them to a running state
func (service *CloudManagementService) restoreTasks() {
	tasks, err := service.dataStore.CloudProvisioning().ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("failed to restore provisioning tasks")
	}

	// First update endpoints that are in provisioning state and have no corresponding provisioning task
	// this can happen with some providers, especially amazon eks
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		log.Error().Err(err).Msg("failed to read environments")
	}

	for _, endpoint := range endpoints {
		if endpoint.Status == portaineree.EndpointStatusProvisioning {
			found := false
			for _, task := range tasks {
				found = task.EndpointID == endpoint.ID
			}

			if !found {
				// Get the associated endpoint and set it's status to error and error detail to timed out
				err := service.setStatus(endpoint.ID, 4)
				if err != nil {
					log.Error().Err(err).Msg("unable to update endpoint status in database")
				}

				summary := "Provisioning Error"
				detail := "Provisioning of this environment has been interrupted and cannot be recovered. This may be due to a Portainer restart. Please check and delete the environment in the cloud platform's portal and remove here."
				err = service.setMessageHandler(endpoint.ID, "")(summary, detail, "error")
				if err != nil {
					log.Error().Err(err).Msg("unable to update endpoint status message in database")
				}
			}
		}
	}

	if len(tasks) > 0 {
		log.Info().Int("count", len(tasks)).Msg("restoring KaaS provisioning tasks")
	}

	for _, task := range tasks {
		if task.CreatedAt.Before(time.Now().Add(-time.Hour * 24 * 7)) {
			log.Info().Msg("removing provisioning task, too old")

			// Get the associated endpoint and set it's status to error and error detail to timed out
			err := service.setStatus(task.EndpointID, 4)
			if err != nil {
				log.Error().Err(err).Msg("unable to update endpoint status in database")
			}

			err = service.setMessageHandler(task.EndpointID, "")("Provisioning Error", "Timed out", "error")
			if err != nil {
				log.Error().Err(err).Msg("unable to update endpoint status message in database")
			}

			// Remove the task from the database because it's too old
			err = service.dataStore.CloudProvisioning().Delete(task.ID)
			if err != nil {
				log.Warn().Err(err).Msg("unable to remove task from the database")
			}

			continue
		}

		// many tasks cannot be restored at the state that they began with,
		// it's safe and reliable to go right back to the beginning
		task.State = int(ProvisioningStatePending)

		_, err := service.dataStore.Endpoint().Endpoint(task.EndpointID)
		if err != nil {
			if service.dataStore.IsErrObjectNotFound(err) {
				if task.Provider == portaineree.CloudProviderKubeConfig {
					log.Info().Int("endpoint_id", int(task.EndpointID)).Msg("removing cluster import task for non-existent endpoint")
				} else {
					log.Info().Int("endpoint_id", int(task.EndpointID)).Msg("removing KaaS provisioning task for non-existent endpoint")
				}

				// Remove the task from the database because it's associated
				// endpoint has been deleted
				err := service.dataStore.CloudProvisioning().Delete(task.ID)
				if err != nil {
					log.Warn().Err(err).Msg("unable to remove task from the database")
				}
			} else {
				log.Error().Err(err).Msg("unable to restore KaaS provisioning task")
			}
		} else {
			go service.provisionKaasClusterTask(task)
		}
	}
}

// changeState changes the state of a task and updates the db
func (service *CloudManagementService) changeState(task *portaineree.CloudProvisioningTask, newState ProvisioningState, message string, operationStatus string) {
	log.Debug().
		Str("cluster_id", task.ClusterID).
		Str("state", newState.String()).
		Msg("changed state of cluster setup task")

	err := service.setMessageHandler(task.EndpointID, "")(message, "", operationStatus)
	if err != nil {
		log.Error().
			Str("cluster_id", task.ClusterID).
			Str("state", ProvisioningState(task.State).String()).
			Err(err).
			Msg("unable to update endpoint status message in database")
	}
	task.State = int(newState)
	task.Retries = 0
}

func (service *CloudManagementService) setMessageHandler(id portaineree.EndpointID, operation string) func(summary, detail, operationStatus string) error {
	return func(summary, detail, operationStatus string) error {
		status := portaineree.EndpointStatusMessage{Summary: summary, Detail: detail, OperationStatus: operationStatus, Operation: operation}
		err := service.dataStore.Endpoint().SetMessage(id, status)
		if err != nil {
			return fmt.Errorf("unable to update endpoint in database")
		}
		return nil
	}
}

func (service *CloudManagementService) setStatus(id portaineree.EndpointID, status int) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(id)
	if err != nil {
		return err
	}

	endpoint.Status = portaineree.EndpointStatus(status)
	err = service.dataStore.Endpoint().UpdateEndpoint(id, endpoint)
	if err != nil {
		return fmt.Errorf("unable to update endpoint in database")
	}
	return nil
}

func (service *CloudManagementService) seedCluster(task *portaineree.CloudProvisioningTask) (err error) {
	customTemplate, err := service.dataStore.CustomTemplate().Read(portaineree.CustomTemplateID(task.CustomTemplateID))
	if err != nil {
		return clouderrors.NewFatalError("error getting custom template with id: %d, error: %v", task.CustomTemplateID, err)
	}

	cluster, err := service.getKaasCluster(task)
	if err != nil {
		return err
	}

	owner, err := service.dataStore.User().Read(task.CreatedByUserID)
	if err != nil {
		return clouderrors.NewFatalError("unable to load user information from the database, error: %v", err)
	}

	namespace, err := kubernetes.GetNamespace([]byte(task.CustomTemplateContent))
	if err != nil || namespace == "" {
		namespace = "default"
	}

	stackID := service.dataStore.Stack().GetNextIdentifier()

	stack := portaineree.Stack{
		ID:              portaineree.StackID(stackID),
		Name:            "seed" + strconv.Itoa(int(task.EndpointID)),
		IsComposeFormat: false,
		Type:            portaineree.KubernetesStack,
		EndpointID:      task.EndpointID,
		Namespace:       namespace,
		EntryPoint:      filesystem.ManifestFileDefaultName,
	}

	stackFolder := strconv.Itoa(stackID)
	projectPath, err := service.fileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(task.CustomTemplateContent))
	if err != nil {
		return clouderrors.NewFatalError("error copying stack manifest for endpoint: %d, error: %v", task.EndpointID, err)
	}
	stack.ProjectPath = projectPath

	err = service.dataStore.Stack().Create(&stack)
	if err != nil {
		return clouderrors.NewFatalError("error creating stack for endpoint: %d, error: %v", task.EndpointID, err)
	}

	labels := kubernetes.KubeAppLabels{
		StackID:   int(stack.ID),
		StackName: customTemplate.Title,
		Owner:     owner.Username,
		Kind:      "content", // should be "content" for custom templates
	}

	labeledManifest, err := kubernetes.AddAppLabels([]byte(task.CustomTemplateContent), labels.ToMap())
	if err != nil {
		return clouderrors.NewFatalError("error adding labels to manifest file: %s, error: %v", task.CustomTemplateContent, err)
	}

	// save the modified manifest to a temp file and deploy it
	manifestFile, err := os.CreateTemp("", "portainer-seed")
	if err != nil {
		return clouderrors.NewFatalError("error creating temp file for manifest, error: %v", err)
	}
	defer os.Remove(manifestFile.Name())

	// write labeledManifest to manifestFile
	_, err = manifestFile.Write(labeledManifest)
	if err != nil {
		return clouderrors.NewFatalError("error writing manifest, error: %v", err)
	}
	manifestFile.Close()

	if task.Provider == portaineree.CloudProviderPreinstalledAgent {
		endpoint, err := service.dataStore.Endpoint().Endpoint(task.EndpointID)
		if err != nil {
			log.Debug().Msgf("error getting endpoint with id: %d, error: %v", task.EndpointID, err)
			return nil
		}

		manifests := []string{manifestFile.Name()}
		_, err = service.kubernetesDeployer.Deploy(
			task.CreatedByUserID,
			endpoint,
			manifests,
			"default",
		)
		return err
	}
	return service.kubernetesDeployer.DeployViaKubeConfig(
		cluster.KubeConfig,
		task.ClusterID,
		manifestFile.Name(),
	)
}

// getKaasCluster gets the kaasCluster object for the task from the associated cloud provider
func (service *CloudManagementService) getKaasCluster(task *portaineree.CloudProvisioningTask) (*KaasCluster, error) {
	endpoint, err := service.dataStore.Endpoint().Endpoint(task.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("unable to read endpoint from the database")
	}

	var credentials *models.CloudCredential
	if task.Provider != portaineree.CloudProviderPreinstalledAgent {
		credentials, err = service.dataStore.CloudCredential().Read(endpoint.CloudProvider.CredentialID)
		if err != nil {
			return nil, fmt.Errorf("unable to read credentials: %w", err)
		}
	}

	cluster := new(KaasCluster)
	switch task.Provider {
	case portaineree.CloudProviderPreinstalledAgent:
		cluster, err = service.PreinstalledAgentGetCluster(task.ClusterID)

	case portaineree.CloudProviderMicrok8s:
		cluster, err = service.Microk8sGetCluster(
			credentials.Credentials["username"],
			credentials.Credentials["password"],
			credentials.Credentials["passphrase"],
			credentials.Credentials["privateKey"],
			task.ClusterID,
			task.MasterNodes[0],
		)

	case portaineree.CloudProviderCivo:
		cluster, err = service.CivoGetCluster(credentials.Credentials["apiKey"], task.ClusterID, task.Region)

	case portaineree.CloudProviderDigitalOcean:
		cluster, err = service.DigitalOceanGetCluster(credentials.Credentials["apiKey"], task.ClusterID)

	case portaineree.CloudProviderLinode:
		cluster, err = service.LinodeGetCluster(credentials.Credentials["apiKey"], task.ClusterID)

	case portaineree.CloudProviderGKE:
		cluster, err = service.GKEGetCluster(credentials.Credentials["jsonKeyBase64"], task.ClusterID, task.Region)

	case portaineree.CloudProviderKubeConfig:
		b64config, ok := credentials.Credentials["kubeconfig"]
		if !ok {
			return nil, fmt.Errorf("failed finding KubeConfig")
		}
		kubeconfig, err := base64.StdEncoding.DecodeString(b64config)
		if err != nil {
			return nil, fmt.Errorf("failed decoding KubeConfig")
		}
		cluster = &KaasCluster{
			Id:         task.ClusterID,
			KubeConfig: string(kubeconfig),
			Ready:      true,
		}

	case portaineree.CloudProviderAzure:
		cluster, err = service.AzureGetCluster(credentials.Credentials, task.ResourceGroup, task.ClusterID)

	case portaineree.CloudProviderAmazon:
		cluster, err = service.AmazonEksGetCluster(credentials.Credentials, task.ClusterID, task.Region)

	default:
		return cluster, fmt.Errorf("%s is not supported", task.Provider)
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to get kaasCluster")
		err = fmt.Errorf("%s is not responding", task.Provider)
	}

	return cluster, err
}

// checkFatal wraps known fatal errors with clouderrors.FatalError which
// shortcuts exiting the privisioning loop.
func checkFatal(err error) error {
	var netError net.Error
	if errors.As(err, &netError) {
		if strings.Contains(err.Error(), "TLS handshake error") ||
			strings.Contains(err.Error(), "connection refused") {
			return clouderrors.NewFatalError(err.Error())
		}
	}

	return err
}

// provisionKaasClusterTask processes a provisioning task
// this function uses a state machine model for progressing the provisioning.  This allows easy retry and state tracking
func (service *CloudManagementService) provisionKaasClusterTask(task portaineree.CloudProvisioningTask) {
	var cluster *KaasCluster
	var kubeClient portaineree.KubeClient
	var serviceIP string
	var err error

	maxAttempts := maxRequestFailures
	if task.Provider == portaineree.CloudProviderKubeConfig {
		maxAttempts = maxRequestFailuresImport
	}

	setMessage := service.setMessageHandler(task.EndpointID, "")
	for {
		var fatal *clouderrors.FatalError
		if errors.As(err, &fatal) {
			task.Err = fatal
		}

		// Handle fatal provisioning errors (such as Quota reached etc)
		// Also handle exceeding max retries in a single state.
		if task.Err != nil || task.Retries >= maxAttempts {
			if task.Err == nil {
				task.Err = err
			}

			service.result <- &cloudPrevisioningResult{
				endpointID: task.EndpointID,
				err:        task.Err,
				state:      int(ProvisioningStateDone),
				taskID:     task.ID,
				provider:   task.Provider,
			}
			return
		}

		switch state := ProvisioningState(task.State); state {
		case ProvisioningStatePending:
			log.Info().
				Str("provider", task.Provider).
				Str("cluster_id", task.ClusterID).
				Int("endpoint_id", int(task.EndpointID)).
				Str("state", state.String()).
				Msg("starting provisionKaasClusterTask")

			switch task.Provider {
			case portaineree.CloudProviderMicrok8s:
				// pendingState logic is completed outside of this function, but this is the initial state
				service.changeState(&task, ProvisioningStateWaitingForCluster, "Waiting for MicroK8s cluster to become available", "processing")
			case portaineree.CloudProviderKubeConfig:
				service.changeState(&task, ProvisioningStateWaitingForCluster, "Importing Kubeconfig", "processing")
			case portaineree.CloudProviderPreinstalledAgent:
				service.changeState(&task, ProvisioningStateDeployingCustomTemplate, "Deploying Custom Template", "processing")
			default:
				service.changeState(&task, ProvisioningStateWaitingForCluster, "Creating KaaS cluster", "processing")
			}

		case ProvisioningStateWaitingForCluster:
			cluster, err = service.getKaasCluster(&task)
			if err != nil {
				task.Retries++
				break
			}

			if cluster.Ready {
				service.changeState(&task, ProvisioningStateAgentSetup, "Deploying Portainer agent", "processing")
			}

			log.Debug().Str("provider", task.Provider).Str("cluster_id", task.ClusterID).Msg("waiting for cluster")

		case ProvisioningStateAgentSetup:
			log.Info().Str("state", state.String()).Int("retries", task.Retries).Msg("process state")

			if kubeClient == nil {
				kubeClient, err = service.clientFactory.CreateKubeClientFromKubeConfig(task.ClusterID, []byte(cluster.KubeConfig))
				if err != nil {
					task.Err = err
					task.Retries++
					break
				}
			}

			log.Info().
				Str("version", kubecli.DefaultAgentVersion).
				Str("provider", task.Provider).
				Str("cluster_id", task.ClusterID).
				Int("endpoint_id", int(task.EndpointID)).
				Msg("deploying Portainer agent")

			// use node port for microk8s
			useNodePort := task.Provider == portaineree.CloudProviderMicrok8s

			err = kubeClient.DeployPortainerAgent(useNodePort)
			if err != nil {
				log.Info().
					Err(err).
					Msg("failed to deploy portainer agent")
				err = checkFatal(err)
				task.Retries++
				break
			}

			service.changeState(&task, ProvisioningStateWaitingForAgent, "Waiting for agent response", "processing")

		case ProvisioningStateWaitingForAgent:
			log.Debug().
				Str("provider", task.Provider).
				Str("cluster_id", task.ClusterID).
				Int("endpoint_id", int(task.EndpointID)).
				Msg("waiting for portainer agent")

			serviceIP, err = kubeClient.GetPortainerAgentAddress(task.MasterNodes)
			if serviceIP == "" {
				detail := "Waiting for the Portainer agent service to be ready (attempt " + strconv.Itoa(task.Retries+1) + " of " + strconv.Itoa(maxAttempts) + ")"
				setMessage("Waiting for agent response", detail, "error")
				if err != nil {
					err = fmt.Errorf("could not get service ip or hostname: %w", err)
				} else {
					err = fmt.Errorf("could not get service ip or hostname")
				}

				log.Debug().
					Err(err).
					Msg("failed to get service ip or hostname")
			}

			if err != nil {
				summary := "Waiting for agent response"
				detail := "Waiting for the Portainer agent service to be ready (attempt " + strconv.Itoa(task.Retries+1) + " of " + strconv.Itoa(maxAttempts) + ")"
				setMessage(summary, detail, "error")
				err = checkFatal(err)
				task.Retries++
				break
			}
			err = kubeClient.CheckRunningPortainerAgentDeployment(task.MasterNodes)
			if err != nil {
				setMessage("Waiting for agent response", "Waiting for the Portainer agent deployment to be ready (attempt "+strconv.Itoa(task.Retries+1)+" of "+strconv.Itoa(maxAttempts)+")", "processing")
				err = checkFatal(err)
				task.Retries++
				break
			}

			log.Debug().
				Str("provider", task.Provider).
				Str("cluster_id", task.ClusterID).
				Str("service_ip", serviceIP).
				Msg("portainer agent service is ready")

			service.changeState(&task, ProvisioningStateUpdatingEnvironment, "Updating environment", "processing")

		case ProvisioningStateUpdatingEnvironment:
			log.Debug().Str("provider", task.Provider).Str("cluster_id", task.ClusterID).Msg("updating environment")
			err = service.updateEndpoint(task.EndpointID, serviceIP)
			if err != nil {
				task.Retries++
				break
			}

			// If custom template is used, we need to deploy custom template
			if task.CustomTemplateID != 0 {
				service.changeState(&task, ProvisioningStateDeployingCustomTemplate, "Deploying Custom Template", "processing")
			} else {
				service.changeState(&task, ProvisioningStateDone, "Connecting", "processing")
			}

		case ProvisioningStateDeployingCustomTemplate:
			if task.CustomTemplateID != 0 {
				log.Debug().Str("provider", task.Provider).Str("cluster_id", task.ClusterID).Msg("deploying custom template")
				err = service.seedCluster(&task)
				if err != nil {
					task.Retries++
					break
				}
			}

			service.changeState(&task, ProvisioningStateDone, "Connecting", "processing")

		case ProvisioningStateDone:
			if err == nil {
				log.Info().
					Str("provider", task.Provider).
					Int("endpoint_id", int(task.EndpointID)).
					Str("cluster_id", task.ClusterID).
					Msg("environment ready")
			}

			service.result <- &cloudPrevisioningResult{
				endpointID: task.EndpointID,
				err:        err,
				state:      int(ProvisioningStateDone),
				taskID:     task.ID,
			}

			return
		}

		// Print state errors and retry counter.
		if err != nil {
			stateStr := ProvisioningState(task.State).String()
			log.Error().Str("state", stateStr).Err(err).Msg("failure in state")

			log.Info().
				Str("state", stateStr).
				Int("attempt", task.Retries).
				Int("max_attempts", maxAttempts).
				Msg("retrying")
		}

		time.Sleep(stateWaitTime)
	}
}

func (service *CloudManagementService) processManagementRequest(request portaineree.CloudManagementRequest) {
	// determine request type
	switch r := request.(type) {
	case *portaineree.CloudProvisioningRequest:
		service.processCreateClusterRequest(r)

	case portaineree.CloudScalingRequest:
		service.processScalingRequest(r)

	case *Microk8sUpdateAddonsRequest:
		service.processMicrok8sUpdateAddonsRequest(r)

	case portaineree.CloudUpgradeRequest:
		service.processClusterUpgradeRequest(r)
	}
}

func (service *CloudManagementService) processClusterUpgradeRequest(request portaineree.CloudUpgradeRequest) {
	var err error
	switch request.Provider() {
	case portaineree.CloudProviderMicrok8s:
		req := request.(*Microk8sUpgradeRequest)
		go func() {
			err = service.processMicrok8sUpgradeRequest(req)
		}()

	default:
		log.Error().Str("provider", request.Provider()).Msg("upgrading not supported for provider")
	}

	if err != nil {
		log.Error().Err(err).Str("provider", request.Provider()).Msg("failed to process cluster upgrade request")
		return
	}

	log.Debug().Err(err).Msg("scaling request complete")
}

func (service *CloudManagementService) processScalingRequest(request portaineree.CloudScalingRequest) {
	var err error
	switch request.Provider() {
	case portaineree.CloudProviderMicrok8s:
		req := request.(*Microk8sScalingRequest)
		go func() {
			err = service.processMicrok8sScalingRequest(req)
		}()

	default:
		log.Error().Str("provider", request.Provider()).Msg("scaling not supported for provider")
	}

	if err != nil {
		log.Error().Err(err).Str("provider", request.Provider()).Msg("failed to process scaling request")
		return
	}

	log.Debug().Err(err).Msg("scaling request complete")
}

func (service *CloudManagementService) processCreateClusterRequest(request *portaineree.CloudProvisioningRequest) {
	log.Info().Str("provider", request.Provider).Str("agent_version", kubecli.DefaultAgentVersion).Msg("new cluster creation request received")

	var credentials *models.CloudCredential
	var err error
	if request.Provider != portaineree.CloudProviderPreinstalledAgent {
		credentials, err = service.dataStore.CloudCredential().Read(request.CredentialID)
		if err != nil {
			log.Error().Err(err).Msg("unable to retrieve credentials from the database")
			return
		}
	}

	var clusterID string
	var provErr error
	// Required for Azure AKS
	var clusterResourceGroup string

	// Note: provErr is logged elsewhere. We just capture it here. Not logging it here avoids
	// it appearing twice in the portainer logs.

	switch request.Provider {
	case portaineree.CloudProviderPreinstalledAgent:
		req := PreinstalledAgentProvisioningClusterRequest{
			EnvironmentID: request.EndpointID,
		}
		clusterID, provErr = service.PreinstalledAgentProvisionCluster(req)

	case portaineree.CloudProviderMicrok8s:
		req := mk8s.Microk8sProvisioningClusterRequest{
			EnvironmentID:     request.EndpointID,
			Credentials:       credentials,
			MasterNodes:       request.MasterNodes,
			WorkerNodes:       request.WorkerNodes,
			Addons:            request.Addons,
			KubernetesVersion: request.KubernetesVersion,
		}
		clusterID, provErr = service.Microk8sProvisionCluster(req)

	case portaineree.CloudProviderCivo:
		req := CloudProvisioningRequest{
			Credentials:       credentials,
			Region:            request.Region,
			ClusterName:       request.Name,
			NodeSize:          request.NodeSize,
			NetworkID:         request.NetworkID,
			NodeCount:         request.NodeCount,
			KubernetesVersion: request.KubernetesVersion,
		}
		clusterID, provErr = service.CivoProvisionCluster(req)

	case portaineree.CloudProviderDigitalOcean:
		req := CloudProvisioningRequest{
			Credentials:       credentials,
			Region:            request.Region,
			ClusterName:       request.Name,
			NodeSize:          request.NodeSize,
			NodeCount:         request.NodeCount,
			KubernetesVersion: request.KubernetesVersion,
		}
		clusterID, provErr = service.DigitalOceanProvisionCluster(req)

	case portaineree.CloudProviderLinode:
		req := CloudProvisioningRequest{
			Credentials:       credentials,
			Region:            request.Region,
			ClusterName:       request.Name,
			NodeSize:          request.NodeSize,
			NodeCount:         request.NodeCount,
			KubernetesVersion: request.KubernetesVersion,
		}
		clusterID, provErr = service.LinodeProvisionCluster(req)

	case portaineree.CloudProviderGKE:
		req := gke.ProvisionRequest{
			APIKey:            credentials.Credentials["jsonKeyBase64"],
			Zone:              request.Region,
			ClusterName:       request.Name,
			Subnet:            request.NetworkID,
			NodeSize:          request.NodeSize,
			CPU:               request.CPU,
			RAM:               request.RAM,
			HDD:               request.HDD,
			NodeCount:         request.NodeCount,
			KubernetesVersion: request.KubernetesVersion,
		}
		clusterID, provErr = service.GKEProvisionCluster(req)

	case portaineree.CloudProviderKubeConfig:
		clusterID = "kubeconfig-" + strconv.Itoa(int(time.Now().Unix()))

	case portaineree.CloudProviderAzure:
		clusterID, clusterResourceGroup, provErr = service.AzureProvisionCluster(credentials.Credentials, request)

	case portaineree.CloudProviderAmazon:
		clusterID, provErr = service.AmazonEksProvisionCluster(credentials.Credentials, request)
	}

	log.Info().Str("provider", request.Provider).Str("cluster-id", clusterID).Msg("creating cluster setup task")

	// Even though there could be a provisioning error above, we still need to continue
	// with the setup task here to properly set the error state of the endpoint
	task, err := service.createClusterSetupTask(request, clusterID, clusterResourceGroup)
	task.Err = provErr
	if err != nil {
		if task.Err == nil {
			// Avoid overwriting previous error. We don't want to give up quite
			// yet because at this point the endpoint exists, but we cannot log
			// our errors to it.
			task.Err = err
		}

		log.Error().Err(err).Msg("failed to create cluster setup task")
	}

	log.Info().Str("provider", request.Provider).Str("cluster-id", clusterID).Msg("provisioning kaas cluster")
	go service.provisionKaasClusterTask(task)
}

func (service *CloudManagementService) processResult(result *cloudPrevisioningResult) {
	log.Info().Msg("cluster creation request completed")

	if result.err != nil {
		log.Error().Err(result.err).Msg("unable to provision cluster")

		err := service.setStatus(result.endpointID, 4)
		if err != nil {
			log.Error().Err(err).Msg("unable to update endpoint status in database")
		}

		// Default error summary.
		if result.errSummary == "" {
			if result.provider == portaineree.CloudProviderKubeConfig {
				result.errSummary = "Connection Error"
			} else {
				result.errSummary = "Provisioning Error"
			}
		}

		err = service.setMessageHandler(result.endpointID, "")(result.errSummary, result.err.Error(), "error")
		if err != nil {
			log.Error().Err(err).Msg("unable to update endpoint status message in database")
		}
	} else {
		err := service.setStatus(result.endpointID, 1)
		if err != nil {
			log.Error().Err(err).Msg("unable to update endpoint status in database")
		}
	}

	// Remove the task from the database
	if result.provider == portaineree.CloudProviderKubeConfig {
		log.Info().Int("endpoint_id", int(result.endpointID)).Msg("removing cluster import task")
	} else {
		log.Info().Int("endpoint_id", int(result.endpointID)).Msg("removing KaaS provisioning task")
	}

	err := service.dataStore.CloudProvisioning().Delete(result.taskID)
	if err != nil {
		log.Error().Err(err).Msg("unable to remove task from the database")
	}
}

func (service *CloudManagementService) updateEndpoint(endpointID portaineree.EndpointID, url string) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return err
	}

	endpoint.URL = url

	err = service.snapshotService.SnapshotEndpoint(endpoint)
	if err != nil {
		return err
	}

	endpoint.SecuritySettings = portaineree.EndpointSecuritySettings{
		AllowVolumeBrowserForRegularUsers: false,
		EnableHostManagementFeatures:      false,

		AllowSysctlSettingForRegularUsers:         true,
		AllowBindMountsForRegularUsers:            true,
		AllowPrivilegedModeForRegularUsers:        true,
		AllowHostNamespaceForRegularUsers:         true,
		AllowContainerCapabilitiesForRegularUsers: true,
		AllowDeviceMappingForRegularUsers:         true,
		AllowStackManagementForRegularUsers:       true,
	}

	err = service.dataStore.Endpoint().UpdateEndpoint(endpointID, endpoint)
	if err != nil {
		return err
	}

	group, err := service.dataStore.EndpointGroup().Read(endpoint.GroupID)
	if err != nil {
		return err
	}

	if len(group.UserAccessPolicies) > 0 || len(group.TeamAccessPolicies) > 0 {
		err = service.authorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return err
		}
	}

	// Run some initial detection. We do not care if these fail, it's just a
	// best effort to enable some optional features if they're supported.
	endpointutils.InitialIngressClassDetection(
		endpoint,
		service.dataStore.Endpoint(),
		service.clientFactory,
	)
	endpointutils.InitialMetricsDetection(
		endpoint,
		service.dataStore.Endpoint(),
		service.clientFactory,
	)
	endpointutils.InitialStorageDetection(
		endpoint,
		service.dataStore.Endpoint(),
		service.clientFactory,
	)

	log.Info().
		Int("endpoint_id", int(endpoint.ID)).
		Str("environment", endpoint.Name).
		Msg("environment successfully created from KaaS cluster")

	return nil
}
