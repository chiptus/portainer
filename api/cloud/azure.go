package cloud

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-02-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/fvbommel/sortorder"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/azure"
	"github.com/portainer/portainer-ee/api/database/models"
	log "github.com/sirupsen/logrus"
)

type (
	pairWithZones struct {
		portaineree.Pair
		Zones []string `json:"zones"`
	}
	nodesByRegion map[string][]pairWithZones
	AzureInfo     struct {
		ResourceGroups     []string                                 `json:"resourceGroups"`
		Regions            []portaineree.Pair                       `json:"regions"`
		NodeSizes          nodesByRegion                            `json:"nodeSizes"`
		KubernetesVersions kubernetesVersions                       `json:"kubernetesVersions"`
		Tier               []containerservice.ManagedClusterSKUTier `json:"tiers"`
	}

	kubernetesVersions []portaineree.Pair
)

func (service *CloudClusterInfoService) AzureGetInfo(credentials *models.CloudCredential) (interface{}, error) {
	if len(credentials.Credentials) == 0 {
		return nil, fmt.Errorf("missing credentials in the database")
	}

	cacheKey := fmt.Sprintf("%s_%d", portaineree.CloudProviderAzure, credentials.ID)
	service.mu.Lock()
	info, ok := service.info[cacheKey]
	service.mu.Unlock()
	var err error
	if !ok {
		// Live fetch if missing cache.
		info, err = service.AzureFetchInfo(credentials.Credentials)
		if err != nil {
			return nil, err
		}
		// Update the cache
		service.mu.Lock()
		service.info[cacheKey] = info
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

func (a *CloudClusterInfoService) AzureFetchInfo(credentials models.CloudCredentialMap) (*AzureInfo, error) {

	log.Debug("[cloud] [message: sending cloud provider info request] [provider: azure]")

	conn, err := azure.NewCredentials(credentials).GetConnection()
	if err != nil {
		return nil, fmt.Errorf("while getting connection: %w", err)
	}

	azureInfo := &AzureInfo{
		Tier: []containerservice.ManagedClusterSKUTier{containerservice.ManagedClusterSKUTierPaid, containerservice.ManagedClusterSKUTierFree},
	}

	// Resource Groups
	groupClient := conn.GetGroupsClient()
	var top int32 = 1000 // TODO: should we add pagination?
	resourceGroups, err := groupClient.ListAll(context.TODO(), "", &top)
	if err != nil {
		return nil, err
	}
	rgs := []string{}
	for _, rg := range resourceGroups {
		// ignore all the resource groups created for AKS by Azure
		// https://docs.microsoft.com/en-us/answers/questions/25725/why-are-two-resource-groups-created-with-aks.html
		if !strings.HasPrefix(*rg.Name, "MC_") {
			rgs = append(rgs, *rg.Name)
		}
	}
	azureInfo.ResourceGroups = rgs

	// Subscription client
	subscriptionClient := conn.GetSubscriptionsClient()

	extendedLocations := false

	// Resource SKUs
	skusClient := conn.GetResourceSkusClient()
	// Pass location filter as blank to get all the skus
	// e.g. location eq 'centralus'
	skus, err := skusClient.List(context.TODO(), "", "true")
	if err != nil {
		return nil, err
	}

	log.Infof("total resource skus [%d]", len(skus.Values()))

	nodesBR := make(nodesByRegion, 0)
	for _, sku := range skus.Values() {
		for _, loc := range *sku.Locations {
			// If resource type is Virtual Machines
			if *sku.ResourceType != "virtualMachines" {
				continue
			}
			location, ok := nodesBR[loc]
			if !ok {
				location = make([]pairWithZones, 0)
			}

			var vCPUs, memory string
			for _, cap := range *sku.Capabilities {
				switch *cap.Name {
				case "vCPUs":
					vCPUs = *cap.Value
				case "MemoryGB":
					memory = *cap.Value
				}
			}
			vcpu, err := strconv.Atoi(vCPUs)
			// vCPUs less than 2 is not supported for Kubernetes
			if err != nil || vcpu < 2 {
				continue
			}

			// Zones
			var zones []string
			for _, l := range *sku.LocationInfo {
				if l.Zones != nil && len(*l.Zones) > 0 {
					zones = *l.Zones
					sort.Strings(zones)
					break
				}
			}

			loc = strings.ToLower(loc)
			name := fmt.Sprintf("%s - %s vCPUs, %sGB Memory", *sku.Name, vCPUs, memory)
			nodesBR[loc] = append(location, pairWithZones{Pair: portaineree.Pair{Value: *sku.Name, Name: name}, Zones: zones})
		}
	}

	// Regions
	locations, err := subscriptionClient.Client.ListLocations(context.TODO(), credentials["subscriptionID"], &extendedLocations)
	if err != nil {
		return nil, err
	}
	log.Infof("number of locations [%d]", len(*locations.Value))
	if len(*locations.Value) == 0 {
		return nil, fmt.Errorf("AKS failed to return any locations")
	}

	locationPairs := make([]portaineree.Pair, 0)
	for _, loc := range *locations.Value {
		// only append regions that have nodes
		if _, ok := nodesBR[*loc.Name]; ok {
			locationPairs = append(locationPairs, portaineree.Pair{Value: *loc.Name, Name: *loc.DisplayName})
		}
	}

	azureInfo.Regions = locationPairs
	firstRegion := locationPairs[0].Value

	// Kube versions
	containerClient := conn.GetContainerServicesClient()
	kubeVersions, err := containerClient.ListKubernetesVersions(context.TODO(), firstRegion)
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(sortorder.Natural(kubeVersions)))
	for _, kv := range kubeVersions {
		azureInfo.KubernetesVersions = append(azureInfo.KubernetesVersions, portaineree.Pair{Name: kv, Value: kv})
	}

	azureInfo.NodeSizes = nodesBR

	return azureInfo, nil
}

func AzureProvisionCluster(credentials models.CloudCredentialMap, params *portaineree.CloudProvisioningRequest) (string, string, error) {
	log.Debugf(
		"[cloud] [message: sending KaaS cluster provisioning request] [provider: azure] [cluster_name: %s] [node_size: %s] [node_count: %d] [region: %s] [tier: %s]",
		params.Name,
		params.NodeSize,
		params.NodeCount,
		params.Region,
		containerservice.ManagedClusterSKUTier(params.Tier),
	)

	conn, err := azure.NewCredentials(credentials).GetConnection()
	if err != nil {
		return "", "", fmt.Errorf("while getting connection: %w", err)
	}

	// Resource Groups
	if params.ResourceGroupName != "" && params.ResourceGroup == "" {
		log.Infof("[cloud] [message: using existing resource group] %s", params.ResourceGroupName)
		groupClient := conn.GetGroupsClient()
		rg, err := groupClient.CreateOrUpdate(context.TODO(), params.ResourceGroupName, resources.Group{
			Name:     &params.ResourceGroupName,
			Location: &params.Region,
		})
		log.Infof("[cloud] [message: using existing resource group] %v", rg.Name)
		if err != nil {
			return "", "", fmt.Errorf("while creating resource group: %w", err)
		}
		params.ResourceGroup = *rg.Name
	}

	containerClient := conn.GetManagedClusterClient()

	trueVar := true
	falseVar := false

	clientID := credentials["clientID"]
	clientSecret := credentials["clientSecret"]

	nodeCountOne := int32(1)

	poolNameSystem := params.PoolName + "s"
	poolNameUser := params.PoolName + "u"

	poolProfiles := []containerservice.ManagedClusterAgentPoolProfile{
		{
			Name:               &poolNameSystem,
			Count:              &nodeCountOne,
			VMSize:             &params.NodeSize,
			OsType:             containerservice.OSTypeLinux,
			Type:               containerservice.AgentPoolTypeVirtualMachineScaleSets,
			EnableNodePublicIP: &falseVar,
			Mode:               containerservice.AgentPoolModeSystem, // explaination: https://docs.microsoft.com/en-us/azure/aks/use-system-pools
			EnableFIPS:         &falseVar,                            // enable when you need Federal Information Processing Standard (FIPS) for security
			AvailabilityZones:  &params.AvailabilityZones,
		},
	}
	if params.NodeCount > 1 {
		nodeCount := int32(params.NodeCount) - nodeCountOne
		poolProfiles = append(poolProfiles, containerservice.ManagedClusterAgentPoolProfile{
			Name:               &poolNameUser,
			Count:              &nodeCount,
			VMSize:             &params.NodeSize,
			OsType:             containerservice.OSTypeLinux,
			Type:               containerservice.AgentPoolTypeVirtualMachineScaleSets,
			EnableNodePublicIP: &falseVar,
			Mode:               containerservice.AgentPoolModeUser,
			EnableFIPS:         &falseVar,
			AvailabilityZones:  &params.AvailabilityZones,
		})
	}

	parameters := containerservice.ManagedCluster{
		Name:     &params.Name,
		Location: &params.Region,
		Sku: &containerservice.ManagedClusterSKU{
			Name: containerservice.ManagedClusterSKUNameBasic,
			Tier: containerservice.ManagedClusterSKUTier(params.Tier),
		},
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			KubernetesVersion: &params.KubernetesVersion,
			DNSPrefix:         &params.DNSPrefix,
			ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfile{
				ClientID: &clientID,
				Secret:   &clientSecret,
			},
			AgentPoolProfiles: &poolProfiles,
			AadProfile: &containerservice.ManagedClusterAADProfile{
				Managed:         &trueVar,
				EnableAzureRBAC: &trueVar,
			},
			EnableRBAC: &trueVar,
			// By default NetworkProfile is set to "Kubenet"
			NetworkProfile: &containerservice.NetworkProfile{NetworkPlugin: containerservice.NetworkPluginKubenet},
		},
	}

	cluster, err := containerClient.CreateOrUpdate(context.TODO(), params.ResourceGroup, params.Name, parameters)
	if err != nil {
		return "", "", err
	}

	var managedCluster containerservice.ManagedCluster
	managedCluster, err = cluster.Result(containerClient.ManagedClustersClient)
	if err != nil {
		log.Warnf("[cloud] [message: error while getting cluster details - returning resourceName as ID] [provider: azure] [cluster_name: %s] [error: %s]", params.Name, err)
		return params.Name, params.ResourceGroup, nil
	}

	return *managedCluster.Name, params.ResourceGroup, nil
}

func AzureGetCluster(credentials models.CloudCredentialMap, resourceGroup, resourceName string) (*KaasCluster, error) {
	conn, err := azure.NewCredentials(credentials).GetConnection()
	if err != nil {
		return nil, fmt.Errorf("while getting connection: %w", err)
	}

	containerClient := conn.GetManagedClusterClient()

	cluster, err := containerClient.Get(context.TODO(), resourceGroup, resourceName)
	if err != nil {
		return nil, fmt.Errorf("while getting cluster: %w", err)
	}

	kaasCluster := &KaasCluster{
		Id:    *cluster.ID,
		Name:  *cluster.Name,
		Ready: false,
	}

	log.Infof("[cloud] [provider: azure] cluster (%s) provisioning status: %s", *cluster.Name, *cluster.ProvisioningState)

	if *cluster.ProvisioningState != "Succeeded" {
		return kaasCluster, fmt.Errorf("cluster is %s", *cluster.ProvisioningState)
	}
	kaasCluster.Ready = true

	if cluster.ManagedClusterProperties != nil && cluster.ManagedClusterProperties.Fqdn != nil {
		log.Infof("[cloud] [provider: azure] cluster (%s) FQDN: (%s)", *cluster.Name, *cluster.ManagedClusterProperties.Fqdn)
		log.Infof("[cloud] [provider: azure] cluster (%s) AzurePortalFQDN (%s)", *cluster.Name, *cluster.ManagedClusterProperties.AzurePortalFQDN)

		adminCreds, err := containerClient.ListClusterAdminCredentials(context.TODO(), resourceGroup, resourceName, *cluster.ManagedClusterProperties.Fqdn)
		if err != nil {
			log.Errorf("[cloud] [message: error while getting cluster admin credentials] [provider: azure] [cluster_name: %s] [error: %s]", resourceName, err)
		}

		for _, c := range *adminCreds.Kubeconfigs {
			if *c.Name == "clusterAdmin" {
				kaasCluster.KubeConfig = string(*c.Value)
				break
			}
		}
	}

	return kaasCluster, nil
}
