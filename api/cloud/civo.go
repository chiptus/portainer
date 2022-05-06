package cloud

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/civo/civogo"
	"github.com/fvbommel/sortorder"
	portaineree "github.com/portainer/portainer-ee/api"
	log "github.com/sirupsen/logrus"
)

type (
	CivoInfo struct {
		Regions            []portaineree.Pair `json:"Regions"`
		NodeSizes          []portaineree.Pair `json:"NodeSizes"`
		Networks           []CivoNetwork      `json:"Networks"`
		KubernetesVersions []string           `json:"KubernetesVersions"`
	}

	CivoNetwork struct {
		Region   string               `json:"Region"`
		Networks []CivoNetworkDetails `json:"Networks"`
	}

	CivoNetworkDetails struct {
		Id   string `json:"Id"`
		Name string `json:"Name"`
	}
)

func (service *CloudClusterInfoService) CivoGetInfo(apiKey string) (interface{}, error) {
	service.mu.Lock()
	info, ok := service.info[portaineree.CloudProviderCivo]
	service.mu.Unlock()
	if !ok {
		// Live fetch if missing cache.
		return service.CivoFetchInfo(apiKey)
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

func (service *CloudClusterInfoService) CivoFetchInfo(apiKey string) (*CivoInfo, error) {
	log.Debug("[cloud] [message: sending cloud provider info request] [provider: civo]")

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

	civoInfo := &CivoInfo{
		Regions:            rs,
		NodeSizes:          ns,
		Networks:           nets,
		KubernetesVersions: kvs,
	}

	// Update the cache also.
	service.mu.Lock()
	service.info[portaineree.CloudProviderCivo] = *civoInfo
	service.mu.Unlock()
	return civoInfo, nil
}

func CivoGetCluster(apiKey, clusterID, region string) (*KaasCluster, error) {
	log.Debugf("[cloud] [message: sending KaaS cluster details request] [provider: civo] [cluster_id: %s] [region: %s]", clusterID, region)

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
	log.Debugf("[cloud] [message: sending KaaS cluster provisioning request] [provider: civo] [cluster_name: %s] [node_size: %s] [node_count: %d] [region: %s]", clusterName, nodeSize, nodeCount, region)

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
