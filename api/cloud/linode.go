package cloud

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"

	"github.com/fvbommel/sortorder"
	"github.com/linode/linodego"
	log "github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type LinodeInfo struct {
	Regions            []portaineree.Pair `json:"regions"`
	NodeSizes          []portaineree.Pair `json:"nodeSizes"`
	KubernetesVersions []portaineree.Pair `json:"kubernetesVersions"`
}

func (service *CloudInfoService) LinodeGetInfo(credential *models.CloudCredential, force bool) (interface{}, error) {
	log.Info().Str("provider", portaineree.CloudProviderLinode).Msg("processing get info request")

	apiKey, ok := credential.Credentials["apiKey"]
	if !ok {
		return nil, errors.New("missing API key in the credentials")
	}

	cacheKey := fmt.Sprintf("%s_%d", portaineree.CloudProviderLinode, credential.ID)

	if force {
		if err := service.linodeFetchRefresh(apiKey, cacheKey); err != nil {
			return nil, err
		}
	}

	service.mu.Lock()
	info, ok := service.info[cacheKey]
	service.mu.Unlock()
	if !ok {
		if err := service.linodeFetchRefresh(apiKey, cacheKey); err != nil {
			return nil, err
		}
	}

	return &info, nil
}

func (service *CloudInfoService) linodeFetchRefresh(apiKey, cacheKey string) error {
	log.Info().Str("provider", portaineree.CloudProviderLinode).Msg("processing fetch info request")

	info, err := service.LinodeFetchInfo(apiKey)
	if err != nil {
		return err
	}

	// Update the cache
	service.mu.Lock()
	service.info[cacheKey] = *info
	service.mu.Unlock()
	return nil
}
func (service *CloudInfoService) LinodeFetchInfo(apiKey string) (*LinodeInfo, error) {
	log.Debug().Str("provider", "linode").Msg("sending cloud provider info request")

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	client := linodego.NewClient(oauth2Client)

	ctx := context.TODO()

	regions, err := client.ListRegions(ctx, &linodego.ListOptions{})
	if err != nil {
		return nil, err
	}

	rs := make([]portaineree.Pair, 0)
	for _, region := range regions {
		r := portaineree.Pair{
			Name:  strings.ToUpper(region.Country) + ": " + region.ID,
			Value: region.ID,
		}

		rs = append(rs, r)
	}

	nodeSizes, err := client.ListTypes(ctx, &linodego.ListOptions{})
	if err != nil {
		return nil, err
	}

	ns := make([]portaineree.Pair, 0)
	for _, size := range nodeSizes {
		// Skip "Nanode 1GB" sized node. It is not valid for a kubernetes
		// deployment (it is not powerful enough) and Linode throws an error
		// when it is selected.
		if size.ID == "g6-nanode-1" {
			continue
		}

		s := portaineree.Pair{
			Name:  size.Label,
			Value: size.ID,
		}

		ns = append(ns, s)
	}

	kubeVersions, err := client.ListLKEVersions(ctx, &linodego.ListOptions{})
	if err != nil {
		return nil, err
	}

	kvs := make([]string, 0)
	for _, version := range kubeVersions {
		kvs = append(kvs, version.ID)
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

	linodeInfo := &LinodeInfo{
		Regions:            rs,
		NodeSizes:          ns,
		KubernetesVersions: versionPairs,
	}

	// Update the cache also.
	service.mu.Lock()
	service.info[portaineree.CloudProviderLinode] = *linodeInfo
	service.mu.Unlock()
	return linodeInfo, nil
}

func (service *CloudManagementService) LinodeGetCluster(apiKey, clusterID string) (*KaasCluster, error) {
	log.Debug().Str("provider", "linode").Str("cluster_id", clusterID).Msg("sending KaaS cluster details request")

	id, err := strconv.Atoi(clusterID)
	if err != nil {
		return nil, err
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	client := linodego.NewClient(oauth2Client)

	ctx := context.TODO()

	kaasCluster := &KaasCluster{
		Id:    clusterID,
		Ready: false,
	}

	kubeConfig, err := client.GetLKEClusterKubeconfig(ctx, id)
	if err != nil {
		return kaasCluster, nil
	}

	kaasCluster.Ready = true

	kubeConfigData, err := base64.StdEncoding.DecodeString(kubeConfig.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed reading kubeconfig %v: %w", kubeConfig.KubeConfig, err)
	}

	kaasCluster.KubeConfig = string(kubeConfigData)

	cluster, err := client.GetLKECluster(ctx, id)
	if err != nil {
		return nil, err
	}

	kaasCluster.Name = cluster.Label

	return kaasCluster, nil
}

func (service *CloudManagementService) LinodeProvisionCluster(req CloudProvisioningRequest) (string, error) {
	log.Debug().
		Str("provider", "linode").
		Str("cluster_name", req.ClusterName).
		Str("node_size", req.NodeSize).
		Int("node_count", req.NodeCount).
		Str("region", req.Region).
		Msg("sending KaaS cluster provisioning request")

	apiKey, ok := req.Credentials.Credentials["apiKey"]
	if !ok {
		return "", errors.New("apiKey not found in credentials")
	}
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	client := linodego.NewClient(oauth2Client)

	// Both DigitalOcean and Civo supports a way to use "latest" but not Linode
	// Also cluster name is lowercased because Linode has a strict validation rule
	clusterConfig := linodego.LKEClusterCreateOptions{
		Label:      strings.ToLower(req.ClusterName),
		Region:     req.Region,
		K8sVersion: req.KubernetesVersion,
		NodePools: []linodego.LKENodePoolCreateOptions{
			{
				Count: req.NodeCount,
				Type:  req.NodeSize,
			},
		},
	}

	ctx := context.TODO()

	cluster, err := client.CreateLKECluster(ctx, clusterConfig)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(cluster.ID), nil
}
