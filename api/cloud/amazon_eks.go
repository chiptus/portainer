package cloud

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/eks"
	"github.com/portainer/portainer-ee/api/database/models"
	log "github.com/sirupsen/logrus"
)

func (service *CloudClusterInfoService) AmazonEksGetInfo(credential *models.CloudCredential) (interface{}, error) {

	cacheKey := fmt.Sprintf("%s_%d", portaineree.CloudProviderAmazon, credential.ID)
	service.mu.Lock()
	info, ok := service.info[cacheKey]
	service.mu.Unlock()
	if !ok {
		// live fetch if missing
		accessKeyId, ok := credential.Credentials["accessKeyId"]
		if !ok {
			return nil, fmt.Errorf("missing EKS accessKeyId for %s (%v)", credential.Name, credential.ID)
		}

		secretAccessKey, ok := credential.Credentials["secretAccessKey"]
		if !ok {
			return nil, fmt.Errorf("missing EKS secretAccessKey for %s (%v)", credential.Name, credential.ID)
		}

		info, err := eks.FetchInfo(accessKeyId, secretAccessKey)
		if err != nil {
			return nil, err
		}

		// Update the cache
		service.mu.Lock()
		service.info[cacheKey] = *info
		service.mu.Unlock()
	}

	// Schedule an update to occur after sending back the cached data. This is
	// needed so the user will get fresh info if they refresh the page twice.
	// For example, if they added a new network to their Civo account, and
	// wanted it to show up without needing to wait 2 hours for the cache to
	// refresh.
	log.Debug("[cloud] [message: used cached cloud data: scheduling new update]")
	service.Update()

	return &info, nil
}

func (service *CloudClusterInfoService) AmazonEksFetchInfo(accessKeyId, secretAccessKey string) (*eks.EksInfo, error) {
	return eks.FetchInfo(accessKeyId, secretAccessKey)
}

func (service *CloudClusterSetupService) AmazonEksGetCluster(credentials models.CloudCredentialMap, name, region string) (*KaasCluster, error) {
	accessKeyId := credentials["accessKeyId"]
	secretAccessKey := credentials["secretAccessKey"]

	prov := eks.NewProvisioner(accessKeyId, secretAccessKey, region, service.fileService.GetKaasFolder())

	cluster, err := prov.GetCluster(name)
	if err != nil {
		return nil, err
	}

	kaasCluster := &KaasCluster{
		Id:    name,
		Name:  name,
		Ready: false,
	}

	if cluster.Status == "ACTIVE" {
		kaasCluster.Ready = true
		kaasCluster.KubeConfig = cluster.KubeConfig
	}

	return kaasCluster, nil
}

func (service *CloudClusterSetupService) AmazonEksProvisionCluster(credentials models.CloudCredentialMap, request *portaineree.CloudProvisioningRequest) (string, error) {
	accessKeyId := credentials["accessKeyId"]
	secretAccessKey := credentials["secretAccessKey"]

	prov := eks.NewProvisioner(accessKeyId, secretAccessKey, request.Region, service.fileService.GetKaasFolder())

	return prov.ProvisionCluster(accessKeyId, secretAccessKey, request.Region, request.Name, request.AmiType, request.InstanceType, request.NodeCount, request.NodeVolumeSize, request.KubernetesVersion)
}
