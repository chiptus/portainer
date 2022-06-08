package cloud

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/gke"
	"github.com/portainer/portainer-ee/api/database/models"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

func (service *CloudClusterInfoService) GKEGetInfo(credential *models.CloudCredential, force bool) (interface{}, error) {
	apiKey, ok := credential.Credentials["jsonKeyBase64"]
	if !ok {
		return nil, errors.New("missing API key for GKE")
	}

	cacheKey := fmt.Sprintf("%s_%d", portaineree.CloudProviderGKE, credential.ID)

	if force {
		if err := service.gkeFetchRefresh(apiKey, cacheKey); err != nil {
			return nil, err
		}
	}

	service.mu.Lock()
	info, ok := service.info[cacheKey]
	service.mu.Unlock()
	if !ok {
		// Live fetch if missing cache.
		if err := service.gkeFetchRefresh(apiKey, cacheKey); err != nil {
			return nil, err
		}
	}

	return &info, nil
}

func (service *CloudClusterInfoService) gkeFetchRefresh(apiKey, cacheKey string) error {
	info, err := service.GKEFetchInfo(apiKey)
	if err != nil {
		return err
	}
	// Update the cache
	service.mu.Lock()
	service.info[cacheKey] = info
	service.mu.Unlock()
	return nil
}

func (service *CloudClusterInfoService) GKEFetchInfo(apiKey string) (*gke.Info, error) {
	log.Debug("[cloud] [message: sending cloud provider info request] [provider: gke]")

	key, err := gke.ExtractKey(apiKey)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	zones, err := key.FetchZones(ctx)
	if err != nil {
		return nil, err
	}

	nets, err := key.FetchNetworks(ctx)
	if err != nil {
		return nil, err
	}

	if len(zones) == 0 {
		return nil, fmt.Errorf("GKE failed to return any regions")
	}
	firstZone := zones[0]

	nodeSizes, err := key.FetchMachines(ctx, firstZone.Value)
	if err != nil {
		return nil, err
	}

	// The min/max and default values for E2-custom are unlikely to change
	// drastically and I haven't found a way to fetch them from google.
	ram := gke.Spec{
		Default: 4,
		Min:     1,
		Max:     16,
	}

	cpu := gke.Spec{
		Default: 2,
		Min:     2,
		Max:     32,
	}

	hdd := gke.Spec{
		Default: 100,
		Min:     10,
		Max:     65536,
	}

	kvs, err := key.FetchVersions(ctx, firstZone.Value)
	if err != nil {
		return nil, err
	}

	gkeInfo := &gke.Info{
		Zones:              zones,
		Networks:           nets,
		NodeSizes:          nodeSizes,
		RAM:                ram,
		CPU:                cpu,
		HDD:                hdd,
		KubernetesVersions: kvs,
	}

	return gkeInfo, nil
}

func GKEGetCluster(apiKey, clusterID, region string) (*KaasCluster, error) {
	log.Debugf("[cloud] [message: sending KaaS cluster details request] [provider: gke] [cluster_id: %s] [region: %s]", clusterID, region)

	key, err := gke.ExtractKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Build the KubeConfig by manually by fetching some information about the
	// cluster.
	ctx := context.Background()
	config, err := key.BuildConfig(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	f, err := os.CreateTemp("", "portainer-gke")
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", f.Name())

	if _, err := f.Write(key.Bytes); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	kaasCluster := &KaasCluster{
		Id:         clusterID,
		Name:       clusterID,
		Ready:      true,
		KubeConfig: string(config),
	}

	return kaasCluster, err
}

func GKEProvisionCluster(r gke.ProvisionRequest) (string, error) {
	log.Debugf("[cloud] [message: sending KaaS cluster provisioning request] [provider: GKE] [cluster_name: %s] [node_count: %d] [zone: %s]", r.ClusterName, r.NodeCount, r.Zone)

	key, err := gke.ExtractKey(r.APIKey)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	containerService, err := container.NewService(
		ctx,
		option.WithCredentialsJSON(key.Bytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed creating container service: %v", err)
	}
	zoneService := container.NewProjectsZonesClustersService(containerService)

	cpuCount := strconv.Itoa(r.CPU)
	memCount := strconv.Itoa(int(math.Round(r.RAM * 1024)))

	machineType := r.NodeSize
	if r.NodeSize == "custom" || r.NodeSize == "" {
		machineType = "e2-custom-" + cpuCount + "-" + memCount
	}
	nodeConfig := container.NodeConfig{
		DiskSizeGb:  int64(r.HDD),
		MachineType: machineType,
	}

	nodePool := container.NodePool{
		// Google's WebUI hardcodes "default-pool" as the default name so I
		// think it makes sense for us to do the same. Currently we don't
		// support multiple node pools.
		Name:             "default-pool",
		InitialNodeCount: int64(r.NodeCount),
		Config:           &nodeConfig,
	}

	// In GKE there are "Networks" and "Subnets". A subnet is part of a network
	// and is associated with a specific region. We only get a subnet from the
	// frontend. Subnets in different networks much have different names. So
	// when we have a subnet name we know there will not be subnets named the
	// same thing in other networks. As a result we can take a subnet and
	// figure out which network it is part of.
	nets, err := key.FetchNetworks(ctx)
	if err != nil {
		return "", err
	}

	network := "default"
	for _, net := range nets {
		for _, sub := range net.Subnets {
			if sub.ID == r.Subnet {
				network = sub.Network
				break
			}
		}
	}

	clusterConfig := container.Cluster{
		Name:                  r.ClusterName,
		Network:               network,
		Subnetwork:            r.Subnet,
		NodePools:             []*container.NodePool{&nodePool},
		InitialClusterVersion: r.KubernetesVersion,
	}

	createReq := &container.CreateClusterRequest{
		Cluster:   &clusterConfig,
		ProjectId: key.ProjectID,
		Zone:      r.Zone,
	}

	_, err = zoneService.Create(key.ProjectID, r.Zone, createReq).Do()
	if err != nil {
		return "", fmt.Errorf("while creating cluster %v: %w", r.ClusterName, err)
	}
	return r.Zone + ":" + r.ClusterName, nil
}
