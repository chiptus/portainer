package cloud

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/eks"
	"github.com/portainer/portainer-ee/api/database/models"
)

func (service *CloudClusterInfoService) AmazonEksGetInfo(credential *models.CloudCredential, force bool) (interface{}, error) {
	cacheKey := fmt.Sprintf("%s_%d", portaineree.CloudProviderAmazon, credential.ID)
	accessKeyId, ok := credential.Credentials["accessKeyId"]
	if !ok {
		return nil, fmt.Errorf("missing EKS accessKeyId for %s (%v)", credential.Name, credential.ID)
	}

	secretAccessKey, ok := credential.Credentials["secretAccessKey"]
	if !ok {
		return nil, fmt.Errorf("missing EKS secretAccessKey for %s (%v)", credential.Name, credential.ID)
	}

	if force {
		if err := service.eksFetchRefresh(accessKeyId, secretAccessKey, cacheKey); err != nil {
			return nil, err
		}
	}

	service.mu.Lock()
	info, ok := service.info[cacheKey]
	service.mu.Unlock()
	if !ok {
		// live fetch if missing
		if err := service.eksFetchRefresh(cacheKey, secretAccessKey, cacheKey); err != nil {
			return nil, err
		}
	}

	return &info, nil
}

func (service *CloudClusterInfoService) eksFetchRefresh(accessKeyId, secretAccessKey, cacheKey string) error {
	info, err := service.AmazonEksFetchInfo(accessKeyId, secretAccessKey)
	if err != nil {
		return err
	}

	// Update the cache
	service.mu.Lock()
	service.info[cacheKey] = *info
	service.mu.Unlock()
	return nil
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
