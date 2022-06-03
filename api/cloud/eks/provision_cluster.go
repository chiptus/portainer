package eks

import (
	"strconv"
	"strings"

	"github.com/portainer/portainer-ee/api/cloud/eks/eksctl"
)

const (
	AmiFamilyAL2          = "AmazonLinux2"
	AmiFamilyBottleRocket = "Bottlerocket"
)

type EksProvisioner struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	BinaryPath      string
}

// Generate EksProvisioner
func NewProvisioner(accessKeyId, secretAccessKey, region, binaryPath string) *EksProvisioner {
	return &EksProvisioner{
		AccessKeyId:     accessKeyId,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		BinaryPath:      binaryPath,
	}
}

func (e *EksProvisioner) ProvisionCluster(accessKeyId, secretAccessKey, region, clusterName, amiType, nodeType string, nodeCount, nodeVolumeSizeGb int, kubernetesVersion string) (string, error) {

	cfg := eksctl.NewConfig(e.AccessKeyId, e.SecretAccessKey, e.Region, e.BinaryPath)

	nodeAmiFamily := AmiFamilyAL2
	if strings.HasPrefix(amiType, "BOTTLEROCKET") {
		nodeAmiFamily = AmiFamilyBottleRocket
	}

	args := []string{"create", "cluster",
		"--name", clusterName,
		"--region", region,
		"--node-ami-family", nodeAmiFamily,
		"--node-type", nodeType,
		"--nodes", strconv.Itoa(nodeCount),
		"--node-volume-size", strconv.Itoa(nodeVolumeSizeGb),
		"--write-kubeconfig=false",
	}

	if kubernetesVersion != "" {
		args = append(args, "--version", kubernetesVersion)
	}

	err := cfg.Run(args...)
	if err != nil {
		return "", err
	}

	return clusterName, nil
}
