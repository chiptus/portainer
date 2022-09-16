package cloud

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"

	"github.com/civo/civogo"
	"github.com/fvbommel/sortorder"
	"github.com/rs/zerolog/log"
)

type (
	CivoInfo struct {
		Regions            []portaineree.Pair `json:"regions"`
		NodeSizes          []portaineree.Pair `json:"nodeSizes"`
		Networks           []CivoNetwork      `json:"networks"`
		KubernetesVersions []portaineree.Pair `json:"kubernetesVersions"`
	}

	CivoNetwork struct {
		Region   string               `json:"region"`
		Networks []CivoNetworkDetails `json:"networks"`
	}

	CivoNetworkDetails struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
)

func (service *CloudClusterInfoService) CivoGetInfo(credential *models.CloudCredential, force bool) (interface{}, error) {
	apiKey, ok := credential.Credentials["apiKey"]
	if !ok {
		return nil, errors.New("missing API key in the credentials")
	}

	cacheKey := fmt.Sprintf("%s_%d", portaineree.CloudProviderCivo, credential.ID)

	if force {
		if err := service.civoFetchRefresh(apiKey, cacheKey); err != nil {
			return nil, err
		}
	}

	service.mu.Lock()
	info, ok := service.info[cacheKey]
	service.mu.Unlock()
	if !ok {
		// Live fetch if missing cache.
		if err := service.civoFetchRefresh(apiKey, cacheKey); err != nil {
			return nil, err
		}
	}

	return &info, nil
}

func (service *CloudClusterInfoService) civoFetchRefresh(apiKey, cacheKey string) error {
	info, err := service.CivoFetchInfo(apiKey)
	if err != nil {
		return err
	}

	// Update the cache
	service.mu.Lock()
	service.info[cacheKey] = *info
	service.mu.Unlock()
	return nil
}

func (service *CloudClusterInfoService) CivoFetchInfo(apiKey string) (*CivoInfo, error) {
	log.Debug().Str("provider", "civo").Msg("sending cloud provider info request")

	client, err := civogo.NewClient(apiKey, "")
	if err != nil {
		return nil, err
	}

	regions, err := client.ListRegions()
	if err != nil {
		return nil, err
	}

	rs := make([]portaineree.Pair, 0)
	nets := make([]CivoNetwork, 0)
	for _, region := range regions {
		if region.Features.Kubernetes {
			r := portaineree.Pair{
				Name:  region.Name,
				Value: region.Code,
			}

			rs = append(rs, r)
		}

		cli, err := civogo.NewClient(apiKey, region.Code)
		if err != nil {
			return nil, err
		}

		networks, err := cli.ListNetworks()
		if err != nil {
			return nil, err
		}

		n := CivoNetwork{
			Region:   region.Code,
			Networks: []CivoNetworkDetails{},
		}

		for _, network := range networks {
			nd := CivoNetworkDetails{
				Id:   network.ID,
				Name: network.Label,
			}

			n.Networks = append(n.Networks, nd)
		}

		nets = append(nets, n)
	}

	nodeSizes, err := client.ListInstanceSizes()
	if err != nil {
		return nil, err
	}

	// We do a few things to clean up the output from Civo's API. Normally,
	// size.Description and size.NiceName are the same strings and will be in
	// the following format:
	// Large - CPU optimized
	// We want to use the CPU optimized part as a prefix so we remove "Large - "
	// from the description string. We also want the name to say "Large" to we
	// remove "- CPU optimized" from the name.
	prefixReg := regexp.MustCompile(`.* - `)
	nameReg := regexp.MustCompile(` - .*`)

	ns := make([]portaineree.Pair, 0)
	for _, size := range nodeSizes {
		// Filter out non-selectable nodes. These seem to be deprecated and may
		// not work for provisioning.
		if !size.Selectable {
			continue
		}

		// Filter out g3 nodes as they are very weak and have issues
		// provisioning. Civo doesn't show them on their own site.
		if strings.HasPrefix(size.Name, "g3") {
			continue
		}

		// Change the description to something like "CPU optimized" by removing
		// the name from the description field. If the Description field doesn't
		// contain a prefix we just set a blank prefix.
		var prefix string
		if !strings.Contains(size.Description, " - ") {
			prefix = ""
		} else {
			prefix = prefixReg.ReplaceAllString(size.Description, "${1}")
			prefix += ": "
		}

		s := portaineree.Pair{
			Name: fmt.Sprintf(
				"%v%v (%vGB RAM - %vGB SSD)",
				prefix,
				nameReg.ReplaceAllString(size.NiceName, "${1}"),
				size.RAMMegabytes/1024,
				size.DiskGigabytes,
			),
			Value: size.Name,
		}

		ns = append(ns, s)
	}

	kubeVersions, err := client.ListAvailableKubernetesVersions()
	if err != nil {
		return nil, err
	}

	kvs := make([]string, 0)
	for _, version := range kubeVersions {
		kvs = append(kvs, version.Version)
	}
	sort.Sort(sort.Reverse(sortorder.Natural(kvs)))

	versionPairs := make([]portaineree.Pair, 0)
	for _, v := range kvs {
		pair := portaineree.Pair{
			Name:  v,
			Value: v,
		}
		versionPairs = append(versionPairs, pair)
	}

	civoInfo := &CivoInfo{
		Regions:            rs,
		NodeSizes:          ns,
		Networks:           nets,
		KubernetesVersions: versionPairs,
	}

	return civoInfo, nil
}

func CivoGetCluster(apiKey, clusterID, region string) (*KaasCluster, error) {
	log.Debug().
		Str("provider", "civo").
		Str("cluster_id", clusterID).
		Str("region", region).
		Msg("sending KaaS cluster details request")

	client, err := civogo.NewClient(apiKey, region)
	if err != nil {
		return nil, err
	}

	cluster, err := client.GetKubernetesCluster(clusterID)
	if err != nil {
		return nil, err
	}

	kaasCluster := &KaasCluster{
		Id:    clusterID,
		Name:  cluster.Name,
		Ready: false,
	}

	if cluster.Status == "ACTIVE" {
		kaasCluster.Ready = true
		kaasCluster.KubeConfig = cluster.KubeConfig
	}

	return kaasCluster, nil
}

func CivoProvisionCluster(apiKey, region, clusterName, nodeSize, networkID string, nodeCount int, kubernetesVersion string) (string, error) {
	log.Debug().
		Str("provider", "civo").
		Str("cluster", clusterName).
		Str("node_size", nodeSize).
		Int("node_count", nodeCount).
		Str("region", region).
		Msg("sending KaaS cluster provisioning request")

	client, err := civogo.NewClient(apiKey, region)
	if err != nil {
		return "", err
	}

	clusterConfig := civogo.KubernetesClusterConfig{
		Name:              clusterName,
		Region:            region,
		NumTargetNodes:    nodeCount,
		TargetNodesSize:   nodeSize,
		NetworkID:         networkID,
		KubernetesVersion: kubernetesVersion,
		FirewallRule:      "all",
	}

	cluster, err := client.NewKubernetesClusters(&clusterConfig)
	if err != nil {
		return "", err
	}

	return cluster.ID, nil
}
