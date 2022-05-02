package cloud

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/fvbommel/sortorder"
	portaineree "github.com/portainer/portainer-ee/api"
	log "github.com/sirupsen/logrus"
)

type DigitalOceanInfo struct {
	Regions            []portaineree.Pair `json:"Regions"`
	NodeSizes          []portaineree.Pair `json:"NodeSizes"`
	KubernetesVersions []string           `json:"KubernetesVersions"`
}

func (service *CloudClusterInfoService) DigitalOceanGetInfo(apiKey string) (interface{}, error) {
	service.mu.Lock()
	info, ok := service.info[portaineree.CloudProviderDigitalOcean]
	service.mu.Unlock()
	if !ok {
		// Live fetch if missing cache.
		return service.DigitalOceanFetchInfo(apiKey)
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

func (service *CloudClusterInfoService) DigitalOceanFetchInfo(apiKey string) (*DigitalOceanInfo, error) {
	log.Debug("[cloud] [message: sending cloud provider info request] [provider: digitalocean]")

	client := godo.NewFromToken(apiKey)

	ctx := context.TODO()

	opts, _, err := client.Kubernetes.GetOptions(ctx)
	if err != nil {
		return nil, err
	}

	rs := make([]portaineree.Pair, 0)
	for _, region := range opts.Regions {
		r := portaineree.Pair{
			Name:  region.Name,
			Value: region.Slug,
		}

		rs = append(rs, r)
	}

	kvs := []string{}
	for _, version := range opts.Versions {
		kvs = append(kvs, version.KubernetesVersion)
	}
	sort.Sort(sort.Reverse(sortorder.Natural(kvs)))
	kvs = append([]string{"latest"}, kvs...)

	nodeSizes, _, err := client.Sizes.List(ctx, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	ns := make([]portaineree.Pair, 0)
	for _, size := range nodeSizes {
		// Skip 1GB nodes as they are not actually valid for Digital Ocean.
		if strings.Contains(size.Slug, "-1gb") {
			continue
		}

		var cpuSuffix string
		if size.Vcpus > 1 {
			cpuSuffix = "CPUs"
		} else {
			cpuSuffix = "CPU"
		}
		cpus := strconv.Itoa(size.Vcpus) + cpuSuffix
		s := portaineree.Pair{
			Name: fmt.Sprintf(
				"%v: (%v - %vGB RAM - %vGB SSD)",
				size.Description,
				cpus,
				size.Memory/1024,
				size.Disk,
			),
			Value: size.Slug,
		}

		ns = append(ns, s)
	}

	digitalOceanInfo := &DigitalOceanInfo{
		Regions:            rs,
		NodeSizes:          ns,
		KubernetesVersions: kvs,
	}

	// Update the cache also.
	service.mu.Lock()
	service.info[portaineree.CloudProviderDigitalOcean] = *digitalOceanInfo
	service.mu.Unlock()
	return digitalOceanInfo, nil
}

func DigitalOceanGetCluster(apiKey, clusterID string) (*KaasCluster, error) {
	log.Debugf("[cloud] [message: sending KaaS cluster details request] [provider: digitalocean] [cluster_id: %s]", clusterID)

	client := godo.NewFromToken(apiKey)

	ctx := context.TODO()

	cluster, _, err := client.Kubernetes.Get(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	kaasCluster := &KaasCluster{
		Id:    clusterID,
		Name:  cluster.Name,
		Ready: false,
	}

	if cluster.Status.State == godo.KubernetesClusterStatusRunning {
		kaasCluster.Ready = true

		config, _, err := client.Kubernetes.GetKubeConfig(ctx, clusterID)
		if err != nil {
			return nil, err
		}

		kaasCluster.KubeConfig = string(config.KubeconfigYAML)
	}

	return kaasCluster, nil
}

func DigitalOceanProvisionCluster(apiKey, region, clusterName, nodeSize string, nodeCount int, kubernetesVersion string) (string, error) {
	log.Debugf("[cloud] [message: sending KaaS cluster provisioning request] [provider: digitalocean] [cluster_name: %s] [node_size: %s] [node_count: %d] [region: %s]", clusterName, nodeSize, nodeCount, region)

	client := godo.NewFromToken(apiKey)

	clusterConfig := godo.KubernetesClusterCreateRequest{
		Name:        strings.ToLower(clusterName),
		RegionSlug:  region,
		VersionSlug: kubernetesVersion,
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			{
				Name:  "default-pool",
				Count: nodeCount,
				Size:  nodeSize,
			},
		},
	}

	ctx := context.TODO()

	cluster, _, err := client.Kubernetes.Create(ctx, &clusterConfig)
	if err != nil {
		return "", err
	}

	return cluster.ID, nil
}
