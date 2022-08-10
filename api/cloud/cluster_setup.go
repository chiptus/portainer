package cloud

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	clouderrors "github.com/portainer/portainer-ee/api/cloud/errors"
	"github.com/portainer/portainer-ee/api/cloud/gke"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
	log "github.com/sirupsen/logrus"
)

type ProvisioningState int

const (
	ProvisioningStatePending ProvisioningState = iota
	ProvisioningStateWaitingForCluster
	ProvisioningStateAgentSetup
	ProvisioningStateWaitingForAgent
	ProvisioningStateUpdatingEndpoint
	ProvisioningStateDone
)

const (
	stateWaitTime      = 15 * time.Second
	maxRequestFailures = 480
)

type (
	CloudClusterSetupService struct {
		dataStore            dataservices.DataStore
		fileService          portaineree.FileService
		shutdownCtx          context.Context
		requests             chan *portaineree.CloudProvisioningRequest
		result               chan *cloudPrevisioningResult
		snapshotService      portaineree.SnapshotService
		authorizationService *authorization.Service
		clientFactory        *kubecli.ClientFactory
	}

	cloudPrevisioningResult struct {
		endpointID portaineree.EndpointID
		state      int
		errSummary string
		err        error
		taskID     portaineree.CloudProvisioningTaskID
		provider   string
	}
)

func NewCloudClusterSetupService(dataStore dataservices.DataStore, fileService portaineree.FileService, clientFactory *kubecli.ClientFactory, snapshotService portaineree.SnapshotService, authorizationService *authorization.Service, shutdownCtx context.Context) *CloudClusterSetupService {
	requests := make(chan *portaineree.CloudProvisioningRequest, 10)
	result := make(chan *cloudPrevisioningResult, 10)

	return &CloudClusterSetupService{
		dataStore:            dataStore,
		fileService:          fileService,
		shutdownCtx:          shutdownCtx,
		requests:             requests,
		result:               result,
		snapshotService:      snapshotService,
		authorizationService: authorizationService,
		clientFactory:        clientFactory,
	}
}

func (service *CloudClusterSetupService) Start() {
	log.Info("[cloud] [message: starting cloud cluster setup service]")

	service.restoreProvisioningTasks()

	go func() {
		for {
			select {
			case request := <-service.requests:
				go service.processRequest(request)

			case result := <-service.result:
				service.processResult(result)

			case <-service.shutdownCtx.Done():
				log.Debugln("[cloud] [message: shutting down KaaS setup queue]")
				return
			}
		}
	}()
}

// Request takes a CloudProvisioningRequest and adds it to the queue to be processed concurrently
func (service *CloudClusterSetupService) Request(r *portaineree.CloudProvisioningRequest) {
	service.requests <- r
}

// createClusterSetupTask transforms a provisioning request into a task and adds it to the db
func (service *CloudClusterSetupService) createClusterSetupTask(request *portaineree.CloudProvisioningRequest, clusterID, resourceGroup string) (portaineree.CloudProvisioningTask, error) {
	task := portaineree.CloudProvisioningTask{
		Provider:      request.Provider,
		ClusterID:     clusterID,
		EndpointID:    request.EndpointID,
		Region:        request.Region,
		State:         request.StartingState,
		CreatedAt:     time.Now(),
		ResourceGroup: resourceGroup,
	}

	return task, service.dataStore.CloudProvisioning().Create(&task)
}

// restoreProvisioningTasks looks up provisioning tasks and retores them to a running state
func (service *CloudClusterSetupService) restoreProvisioningTasks() {
	tasks, err := service.dataStore.CloudProvisioning().Tasks()
	if err != nil {
		log.Errorf("[cloud] [message: failed to restore provisioning tasks] [error: %v]", err)
	}

	// First update endpoints that are in provisioning state and have no corresponding provisioning task
	// this can happen with some providers, especially amazon eks
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		log.Errorf("[cloud] [message: failed to read environments] [error: %v]", err)
	}

	for _, endpoint := range endpoints {
		if endpoint.Status == portaineree.EndpointStatusProvisioning {
			found := false
			for _, task := range tasks {
				found = task.EndpointID == task.EndpointID
			}

			if !found {
				// Get the associated endpoint and set it's status to error and error detail to timed out
				err := service.setStatus(endpoint.ID, 4)
				if err != nil {
					log.Errorf("[cloud] [message: unable to update endpoint status in database] [error: %v]", err)
				}

				err = service.setMessage(endpoint.ID, "Provisioning Error", "Provisioning of this environment has been interrupted and cannot be recovered. This may be due to a Portainer restart. Please check and delete the environment in the cloud platform's portal and remove here.")
				if err != nil {
					log.Errorf("[cloud] [message: unable to update endpoint status message in database] [error: %v]", err)
				}
			}
		}
	}

	if len(tasks) > 0 {
		log.Infof("[cloud] [message: restoring %d KaaS provisioning tasks]", len(tasks))
	}

	for _, task := range tasks {
		if task.CreatedAt.Before(time.Now().Add(-time.Hour * 24 * 7)) {
			log.Infof("[cloud] [message: removing provisioning task, too old]")

			// Get the associated endpoint and set it's status to error and error detail to timed out
			err := service.setStatus(task.EndpointID, 4)
			if err != nil {
				log.Errorf("[cloud] [message: unable to update endpoint status in database] [error: %v]", err)
			}

			err = service.setMessage(task.EndpointID, "Provisioning Error", "Timed out")
			if err != nil {
				log.Errorf("[cloud] [message: unable to update endpoint status message in database] [error: %v]", err)
			}

			// Remove the task from the database because it's too old
			err = service.dataStore.CloudProvisioning().Delete(task.ID)
			if err != nil {
				log.Warnf("[cloud] [message: unable to remove task from the database] [error: %v]", err)
			}
			continue
		}

		// many tasks cannot be restored at the state that they began with,
		// it's safe and reliable to go right back to the beginning
		task.State = int(ProvisioningStatePending)

		_, err := service.dataStore.Endpoint().Endpoint(task.EndpointID)
		if err != nil {
			if service.dataStore.IsErrObjectNotFound(err) {
				log.Infof("[cloud] [message: removing KaaS provisioning task for non-existent endpoint] [endpointID: %d]", task.EndpointID)

				// Remove the task from the database because it's associated
				// endpoint has been deleted
				err := service.dataStore.CloudProvisioning().Delete(task.ID)
				if err != nil {
					log.Warnf("[cloud] [message: unable to remove task from the database] [error: %v]", err)
				}
			} else {
				log.Errorf("[cloud] [message: unable to restore KaaS provisioning task] [error: %v]", err)
			}
		} else {
			go service.provisionKaasClusterTask(task)
		}
	}
}

// changeState changes the state of a task and updates the db
func (service *CloudClusterSetupService) changeState(task *portaineree.CloudProvisioningTask, newState ProvisioningState, message string) {
	log.Debugf("[cloud] [message: changed state of cluster setup task] [clusterID: %s] [state: %s]", task.ClusterID, newState)
	err := service.setMessage(task.EndpointID, message, "")
	if err != nil {
		log.Errorf("[cloud] [message: unable to update endpoint status message in database] [clusterID: %s] [state: %s] [error: %v]", task.ClusterID, ProvisioningState(task.State), err)
	}
	task.State = int(newState)
	task.Retries = 0
}

func (service *CloudClusterSetupService) setMessage(id portaineree.EndpointID, summary string, detail string) error {
	endpoint, err := service.dataStore.Endpoint().Endpoint(id)
	if err != nil {
		return err
	}

	endpoint.StatusMessage.Summary = summary
	endpoint.StatusMessage.Detail = detail
	err = service.dataStore.Endpoint().UpdateEndpoint(id, endpoint)
	if err != nil {
		return fmt.Errorf("unable to update endpoint in database")
	}
	return nil
}

func (service *CloudClusterSetupService) setStatus(id portaineree.EndpointID, status int) error {
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

// getKaasCluster gets the kaasCluster object for the task from the associated cloud provider
func (service *CloudClusterSetupService) getKaasCluster(task *portaineree.CloudProvisioningTask) (cluster *KaasCluster, err error) {
	endpoint, err := service.dataStore.Endpoint().Endpoint(task.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("unable to read endpoint from the database")
	}

	credentials, err := service.dataStore.CloudCredential().GetByID(endpoint.CloudProvider.CredentialID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read credentials from the database")
	}

	switch task.Provider {
	case portaineree.CloudProviderCivo:
		cluster, err = CivoGetCluster(credentials.Credentials["apiKey"], task.ClusterID, task.Region)

	case portaineree.CloudProviderDigitalOcean:
		cluster, err = DigitalOceanGetCluster(credentials.Credentials["apiKey"], task.ClusterID)

	case portaineree.CloudProviderLinode:
		cluster, err = LinodeGetCluster(credentials.Credentials["apiKey"], task.ClusterID)

	case portaineree.CloudProviderGKE:
		cluster, err = GKEGetCluster(credentials.Credentials["jsonKeyBase64"], task.ClusterID, task.Region)

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
		cluster, err = AzureGetCluster(credentials.Credentials, task.ResourceGroup, task.ClusterID)

	case portaineree.CloudProviderAmazon:
		cluster, err = service.AmazonEksGetCluster(credentials.Credentials, task.ClusterID, task.Region)

	default:
		return cluster, fmt.Errorf("%s is not supported", task.Provider)
	}

	if err != nil {
		log.Errorf("[cloud] [message: failed to get kaasCluster: %v]", err)
		err = fmt.Errorf("%s is not responding", task.Provider)
	}
	return cluster, err
}

// Wraps any fatal network error with FatalError type from clouderrors
// which shortcuts exiting the privisioning loop
func checkFatal(err error) error {
	if err != nil {
		if _, ok := err.(net.Error); ok {
			if strings.Contains(err.Error(), "TLS handshake error") ||
				strings.Contains(err.Error(), "connection refused") {
				return clouderrors.NewFatalError(err.Error())
			}
		}
	}

	return nil
}

// provisionKaasClusterTask processes a provisioning task
// this function uses a state machine model for progressing the provisioning.  This allows easy retry and state tracking
func (service *CloudClusterSetupService) provisionKaasClusterTask(task portaineree.CloudProvisioningTask) {
	var cluster *KaasCluster
	var kubeClient portaineree.KubeClient
	var serviceIP string
	var err error

	for {
		var fatal *clouderrors.FatalError
		if errors.As(err, &fatal) {
			task.Err = fatal
		}

		// Handle fatal provisioning errors (such as Quota reached etc)
		// Also handle exceeding max retries in a single state.
		if task.Err != nil || task.Retries >= maxRequestFailures {
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

		switch ProvisioningState(task.State) {
		case ProvisioningStatePending:
			log.Infof("[message: starting provisionKaasClusterTask] [provider: %s] [clusterId: %s] [endpointId: %d] [state: %d]", task.Provider, task.ClusterID, task.EndpointID, task.State)

			msg := "Creating KaaS cluster"
			if task.Provider == portaineree.CloudProviderKubeConfig {
				msg = "Importing Kubeconfig"
			}

			// pendingState logic is completed outside of this function, but this is the initial state
			service.changeState(&task, ProvisioningStateWaitingForCluster, msg)

		case ProvisioningStateWaitingForCluster:
			cluster, err = service.getKaasCluster(&task)
			if err != nil {
				task.Retries++
				break
			}

			if cluster.Ready {
				service.changeState(&task, ProvisioningStateAgentSetup, "Deploying Portainer agent")
			}

			log.Debugf("[message: waiting for cluster] [provider: %s] [clusterId: %s]", task.Provider, task.ClusterID)

		case ProvisioningStateAgentSetup:
			log.Infof("[message: process state] [state: %s] [retries: %d]", ProvisioningState(task.State), task.Retries)

			if kubeClient == nil {
				kubeClient, err = service.clientFactory.CreateKubeClientFromKubeConfig(task.ClusterID, cluster.KubeConfig)
				if err != nil {
					task.Err = err
					task.Retries++
					break
				}
			}

			log.Infof("[message: checking for old portainer namespace] [state: %s] [retries: %d]", ProvisioningState(task.State), task.Retries)
			err = kubeClient.DeletePortainerAgent()
			if err != nil {
				err = checkFatal(err)
				task.Retries++
				break
			}

			log.Infof("[message: deploying Portainer agent version: %s] [provider: %s] [clusterId: %s] [endpointId: %d]", kubecli.DefaultAgentVersion, task.Provider, task.ClusterID, task.EndpointID)
			err = kubeClient.DeployPortainerAgent()
			if err != nil {
				err = checkFatal(err)
				task.Retries++
				break
			}

			service.changeState(&task, ProvisioningStateWaitingForAgent, "Waiting for agent response")

		case ProvisioningStateWaitingForAgent:
			log.Debugf("[message: waiting for portainer agent] [provider: %s] [clusterId: %s] [endpointId: %d]", task.Provider, task.ClusterID, task.EndpointID)

			serviceIP, err = kubeClient.GetPortainerAgentIPOrHostname()
			if serviceIP == "" {
				err = fmt.Errorf("could not get service ip or hostname: %v", err)
			}

			if err != nil {
				err = checkFatal(err)
				task.Retries++
				break
			}

			log.Debugf("[cloud] [message: portainer agent service is ready] [provider: %s] [clusterId:%s] [serviceIP: %s]", task.Provider, task.ClusterID, serviceIP)
			service.changeState(&task, ProvisioningStateUpdatingEndpoint, "Updating environment")

		case ProvisioningStateUpdatingEndpoint:
			log.Debugf("[message: updating environment] [provider: %s] [clusterId: %s]", task.Provider, task.ClusterID)
			err = service.updateEndpoint(task.EndpointID, fmt.Sprintf("%s:9001", serviceIP))
			if err != nil {
				task.Retries++
				break
			}

			service.changeState(&task, ProvisioningStateDone, "Connecting")

		case ProvisioningStateDone:
			if err != nil {
				log.Infof("[message: environment ready] [provider: %s] [endpointId: %d] [clusterId: %s]", task.Provider, task.EndpointID, task.ClusterID)
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
			log.Errorf("[cloud] [message: failure in state %s] [err: %v]", ProvisioningState(task.State), err)
			log.Infof("[cloud] [message: retrying] [state: %s] [attempt: %d of %d]", ProvisioningState(task.State), task.Retries, maxRequestFailures)
		}

		time.Sleep(stateWaitTime)
	}
}

func (service *CloudClusterSetupService) processRequest(request *portaineree.CloudProvisioningRequest) {
	log.Info("[cloud] [message: new cluster creation request received]")
	log.Infof("[cloud] [message: will use agent version: %s]", kubecli.DefaultAgentVersion)

	credentials, err := service.dataStore.CloudCredential().GetByID(request.CredentialID)
	if err != nil {
		log.Errorf("[cloud] [message: unable to retrieve credentials from the database] [error: %v]", err)
		return
	}

	var clusterID string
	var provErr error
	// Required for Azure AKS
	var clusterResourceGroup string

	// Note: provErr is logged elsewhere. We just capture it here. Not logging it here avoids
	// it appearing twice in the portainer logs.

	switch request.Provider {
	case portaineree.CloudProviderCivo:
		clusterID, provErr = CivoProvisionCluster(credentials.Credentials["apiKey"], request.Region, request.Name, request.NodeSize, request.NetworkID, request.NodeCount, request.KubernetesVersion)

	case portaineree.CloudProviderDigitalOcean:
		clusterID, provErr = DigitalOceanProvisionCluster(credentials.Credentials["apiKey"], request.Region, request.Name, request.NodeSize, request.NodeCount, request.KubernetesVersion)

	case portaineree.CloudProviderLinode:
		clusterID, provErr = LinodeProvisionCluster(credentials.Credentials["apiKey"], request.Region, request.Name, request.NodeSize, request.NodeCount, request.KubernetesVersion)

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
		clusterID, provErr = GKEProvisionCluster(req)

	case portaineree.CloudProviderKubeConfig:
		clusterID = "kubeconfig-" + strconv.Itoa(int(time.Now().Unix()))

	case portaineree.CloudProviderAzure:
		clusterID, clusterResourceGroup, provErr = AzureProvisionCluster(credentials.Credentials, request)

	case portaineree.CloudProviderAmazon:
		clusterID, provErr = service.AmazonEksProvisionCluster(credentials.Credentials, request)
	}

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
		log.Errorf("Failed to create cluster setup task %v", err)
	}

	go service.provisionKaasClusterTask(task)
}

func (service *CloudClusterSetupService) processResult(result *cloudPrevisioningResult) {
	log.Info("[cloud] [message: cluster creation request completed]")

	if result.err != nil {
		log.Errorf("[cloud] [message: unable to provision cluster] [error: %v]", result.err)
		err := service.setStatus(result.endpointID, 4)
		if err != nil {
			log.Errorf("[cloud] [message: unable to update endpoint status in database] [error: %v]", err)
		}

		// Default error summary.
		if result.errSummary == "" {
			if result.provider == portaineree.CloudProviderKubeConfig {
				result.errSummary = "Connection Error"
			} else {
				result.errSummary = "Provisioning Error"
			}
		}

		err = service.setMessage(result.endpointID, result.errSummary, result.err.Error())
		if err != nil {
			log.Errorf("[cloud] [message: unable to update endpoint status message in database] [error: %v]", err)
		}
	}

	// Remove the task from the database
	log.Infof("[cloud] [message: removing KaaS provisioning task] [endpointID: %d]", result.endpointID)

	err := service.dataStore.CloudProvisioning().Delete(result.taskID)
	if err != nil {
		log.Errorf("[cloud] [message: unable to remove task from the database] [error: %v]", err)
	}
}

func (service *CloudClusterSetupService) updateEndpoint(endpointID portaineree.EndpointID, url string) error {
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

	group, err := service.dataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
	if err != nil {
		return err
	}

	if len(group.UserAccessPolicies) > 0 || len(group.TeamAccessPolicies) > 0 {
		err = service.authorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return err
		}
	}

	log.Infof("[cloud] [message: environment successfully created from KaaS cluster] [endpointId: %d] [environment: %s]", endpoint.ID, endpoint.Name)
	return nil
}
