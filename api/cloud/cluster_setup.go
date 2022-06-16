package cloud

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
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
				service.processRequest(request)

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
		log.Errorf("[cloud] [message: failed to restore provisioning tasks] [error: %s]", err)
	}

	if len(tasks) > 0 {
		log.Infof("[cloud] [message: restoring %d KaaS provisioning tasks]", len(tasks))
	}

	for _, task := range tasks {
		if task.CreatedAt.Before(time.Now().Add(-time.Hour * 24 * 7)) {
			log.Infof("[cloud] [message: removing provisioning task, too old]")

			// Remove the task from the database because it's too old
			err := service.dataStore.CloudProvisioning().Delete(task.ID)
			if err != nil {
				log.Warnf("[cloud] [message: unable to remove task from the database] [error: %s]", err.Error())
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
					log.Warnf("[cloud] [message: unable to remove task from the database] [error: %s]", err.Error())
				}
			} else {
				log.Errorf("[cloud] [message: unable to restore KaaS provisioning task] [error: %s]", err.Error())
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
		return cluster, fmt.Errorf("%v is not supported", task.Provider)
	}

	if err != nil {
		log.Errorf("[cloud] [message: failed to get kaasCluster: %v]", err)
		err = fmt.Errorf("%v is not responding", task.Provider)
	}
	return cluster, err
}

// provisionKaasClusterTask processes a provisioning task
// this function uses a state machine model for progressing the provisioning.  This allows easy retry and state tracking
func (service *CloudClusterSetupService) provisionKaasClusterTask(task portaineree.CloudProvisioningTask) {
	var cluster *KaasCluster
	var kubeClient portaineree.KubeClient
	var serviceIP string
	var err error
	var fatal bool = false

	log.Infof("[message: starting provisionKaasClusterTask] [provider: %s] [clusterId: %s] [endpointId: %d] [state: %d]", task.Provider, task.ClusterID, task.EndpointID, task.State)
	for {
		switch ProvisioningState(task.State) {
		case ProvisioningStatePending:
			// pendingState logic is completed outside of this function, but this is the initial state
			service.changeState(&task, ProvisioningStateWaitingForCluster, "Creating KaaS cluster")

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
					// If KubeClient is not created from KubeConfig, it means
					// KubeConfig is incorrect
					fatal = true
					task.Retries++
					break
				}
			}

			log.Infof("[message: checking for old portainer namespace] [state: %s] [retries: %d]", ProvisioningState(task.State), task.Retries)
			err = kubeClient.DeletePortainerAgent()
			if err != nil {
				task.Retries++
				break
			}

			log.Infof("[message: deploying Portainer agent version: %s] [provider: %s] [clusterId: %s] [endpointId: %d]", kubecli.DefaultAgentVersion, task.Provider, task.ClusterID, task.EndpointID)
			err = kubeClient.DeployPortainerAgent()
			if err != nil {
				task.Retries++
				break
			}

			service.changeState(&task, ProvisioningStateWaitingForAgent, "Waiting for agent response")

		case ProvisioningStateWaitingForAgent:
			log.Debugf("[message: waiting for portainer agent] [provider: %s] [clusterId: %s] [endpointId: %d]", task.Provider, task.ClusterID, task.EndpointID)

			serviceIP, err = kubeClient.GetPortainerAgentIPOrHostname()
			if serviceIP == "" {
				err = fmt.Errorf("have not recieved agent service ip yet")
			}

			if err != nil {
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

		if _, ok := err.(clouderrors.FatalError); ok {
			fatal = true
		}

		// Handle exceeding max retries in a single state.
		if task.Retries >= maxRequestFailures || fatal {
			service.result <- &cloudPrevisioningResult{
				endpointID: task.EndpointID,
				err:        err,
				state:      int(ProvisioningStateDone),
				taskID:     task.ID,
			}
			return
		}

		// Handle initial fatal provisioning errors (such as Quota reached or a
		// broken datastore).
		if task.Err != nil {
			service.result <- &cloudPrevisioningResult{
				endpointID: task.EndpointID,
				err:        task.Err,
				state:      int(ProvisioningStateDone),
				taskID:     task.ID,
			}
			return
		}

		time.Sleep(stateWaitTime)
	}
}

func (service *CloudClusterSetupService) processRequest(request *portaineree.CloudProvisioningRequest) {
	log.Info("[cloud] [message: new cluster creation request received]")
	log.Infof("[cloud] [message: will use agent version: %s]", kubecli.DefaultAgentVersion)

	credentials, err := service.dataStore.CloudCredential().GetByID(request.CredentialID)
	if err != nil {
		log.Errorf("[cloud] [message: unable to retrieve credentials from the database] [error: %v]", err.Error())
		return
	}

	var clusterID string
	var provErr error
	// Required for Azure AKS
	var clusterResourceGroup string

	switch request.Provider {
	case portaineree.CloudProviderCivo:
		clusterID, provErr = CivoProvisionCluster(credentials.Credentials["apiKey"], request.Region, request.Name, request.NodeSize, request.NetworkID, request.NodeCount, request.KubernetesVersion)
		if provErr != nil {
			log.Errorf("[cloud] [message: Civo cluster provisioning failed %v]", provErr)
		}

	case portaineree.CloudProviderDigitalOcean:
		clusterID, provErr = DigitalOceanProvisionCluster(credentials.Credentials["apiKey"], request.Region, request.Name, request.NodeSize, request.NodeCount, request.KubernetesVersion)
		if provErr != nil {
			log.Errorf("[cloud] [message: Digital Ocean cluster provisioning failed %v]", provErr)
		}

	case portaineree.CloudProviderLinode:
		clusterID, provErr = LinodeProvisionCluster(credentials.Credentials["apiKey"], request.Region, request.Name, request.NodeSize, request.NodeCount, request.KubernetesVersion)
		if provErr != nil {
			log.Errorf("[cloud] [message: Linode cluster provisioning failed %v]", provErr)
		}

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
		if provErr != nil {
			log.Errorf("[cloud] [message: GKE cluster provisioning failed %v]", provErr)
		}

	case portaineree.CloudProviderKubeConfig:
		clusterID = "kubeconfig-" + strconv.Itoa(int(time.Now().Unix()))

	case portaineree.CloudProviderAzure:
		clusterID, clusterResourceGroup, provErr = AzureProvisionCluster(credentials.Credentials, request)
		if provErr != nil {
			log.Errorf("[cloud] [message: Azure cluster provisioning failed %v]", provErr)
		}

	case portaineree.CloudProviderAmazon:
		clusterID, provErr = service.AmazonEksProvisionCluster(credentials.Credentials, request)
		if provErr != nil {
			log.Errorf("[cloud] [message: Amazon cluster provisioning failed %v]", provErr)
		}
	}

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
		log.Errorf("[cloud] [message: unable to provision cluster] [error: %v]", result.err.Error())
		err := service.setStatus(result.endpointID, 4)
		if err != nil {
			log.Errorf("[cloud] [message: unable to update endpoint status in database] [error: %v]", err)
		}

		// Default error summary.
		if result.errSummary == "" {
			result.errSummary = "Provisioning Error"
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
		log.Errorf("[cloud] [message: unable to remove task from the database] [error: %v]", err.Error())
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
