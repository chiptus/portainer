package cloud

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/azure"
	clouderrors "github.com/portainer/portainer-ee/api/cloud/errors"
	"github.com/portainer/portainer-ee/api/database/models"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-02-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/fvbommel/sortorder"
	"github.com/rs/zerolog/log"
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

func (service *CloudClusterInfoService) AzureGetInfo(credentials *models.CloudCredential, force bool) (interface{}, error) {
	if len(credentials.Credentials) == 0 {
		return nil, fmt.Errorf("missing credentials in the database")
	}

	cacheKey := fmt.Sprintf("%s_%d", portaineree.CloudProviderAzure, credentials.ID)

	if force {
		if err := service.azureFetchRefresh(credentials, cacheKey); err != nil {
			return nil, err
		}
	}

	service.mu.Lock()
	info, ok := service.info[cacheKey]
	service.mu.Unlock()
	if !ok {
		// Live fetch if missing cache.
		if err := service.azureFetchRefresh(credentials, cacheKey); err != nil {
			return nil, err
		}
	}

	return &info, nil
}

func (service *CloudClusterInfoService) azureFetchRefresh(c *models.CloudCredential, cacheKey string) error {
	info, err := service.AzureFetchInfo(c.Credentials)
	if err != nil {
		return err
	}

	// Update the cache
	service.mu.Lock()
	service.info[cacheKey] = info
	service.mu.Unlock()
	return nil
}

func (a *CloudClusterInfoService) AzureFetchInfo(credentials models.CloudCredentialMap) (*AzureInfo, error) {
	log.Debug().Msg("sending cloud provider info request")

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

	log.Info().Int("count", len(skus.Values())).Msg("total resource skus")

	nodesBR := make(nodesByRegion, 0)
	for _, sku := range skus.Values() {
		for _, loc := range *sku.Locations {
			// Skip when resource type is not Virtual Machines or family is basic
			if *sku.ResourceType != "virtualMachines" || *sku.Family == "basicAFamily" {
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

	log.Info().Int("count", len(*locations.Value)).Msg("number of locations")
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
	log.Debug().
		Str("provider", "azure").
		Str("cluster_name", params.Name).
		Str("node_size", params.NodeSize).
		Int("node_count", params.NodeCount).
		Str("region", params.Region).
		Str("tier", string(containerservice.ManagedClusterSKUTier(params.Tier))).
		Msg("sending KaaS cluster provisioning request")

	conn, err := azure.NewCredentials(credentials).GetConnection()
	if err != nil {
		return "", "", fmt.Errorf("while getting connection: %w", err)
	}

	// Resource Groups
	if params.ResourceGroupName != "" && params.ResourceGroup == "" {
		log.Info().Str("resource_group", params.ResourceGroupName).Msg("using existing resource group")

		groupClient := conn.GetGroupsClient()
		rg, err := groupClient.CreateOrUpdate(context.TODO(), params.ResourceGroupName, resources.Group{
			Name:     &params.ResourceGroupName,
			Location: &params.Region,
		})

		log.Info().Str("resource_group", *rg.Name).Msg("using existing resource group")

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
		log.Warn().
			Str("provider", "azure").
			Str("cluster_name", params.Name).
			Err(err).
			Msg("error while getting cluster details, returning resourceName as ID")

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

	log.Info().
		Str("provider", "azure").
		Str("cluster", *cluster.Name).
		Str("status", *cluster.ProvisioningState).
		Msg("cluster provisioning")

	if *cluster.ProvisioningState != "Succeeded" {
		return kaasCluster, fmt.Errorf("cluster is %s", *cluster.ProvisioningState)
	}

	if cluster.ManagedClusterProperties != nil && cluster.ManagedClusterProperties.Fqdn != nil {
		log.Info().
			Str("cluster", *cluster.Name).
			Str("FQDN", *cluster.ManagedClusterProperties.Fqdn).
			Str("azure_portal_FQDN", *cluster.ManagedClusterProperties.AzurePortalFQDN).
			Msg("")

		adminCreds, err := containerClient.ListClusterAdminCredentials(context.TODO(), resourceGroup, resourceName, *cluster.ManagedClusterProperties.Fqdn)
		if err != nil {
			log.Error().
				Str("provider", "azure").
				Str("cluster", *cluster.Name).
				Err(err).
				Msg("error while getting cluster admin credentials")

			return nil, clouderrors.NewFatalError("error while getting azure cluster admin credentials for resource (%s), err: %s", resourceName, err.Error())
		}

		for _, c := range *adminCreds.Kubeconfigs {
			if *c.Name == "clusterAdmin" {
				kaasCluster.KubeConfig = string(*c.Value)
				break
			}
		}
	}
	kaasCluster.Ready = true

	return kaasCluster, nil
}
