package eks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/fvbommel/sortorder"
	portaineree "github.com/portainer/portainer-ee/api"
)

const defaultRegion = "us-west-2"

type (
	InstanceType struct {
		portaineree.Pair
		CompatibleAmiTypes []string `json:"compatibleAmiTypes" example:"AL2_x86_64"`
	}

	EksInfo struct {
		Regions            []portaineree.Pair        `json:"regions"`
		KubernetesVersions []portaineree.Pair        `json:"kubernetesVersions"`
		AmiTypes           []portaineree.Pair        `json:"amiTypes"`
		InstanceTypes      map[string][]InstanceType `json:"instanceTypes"`
	}
)

func loadConfig(accessKeyId, secretAccessKey string, region *string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithDefaultRegion(defaultRegion), // we always need a default region otherwise we can't query the AWS API
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	)

	if region != nil {
		// if a region is provided we override the default
		cfg.Region = *region
	}

	return cfg, err
}

type InstanceTypeByName []InstanceType

func (t InstanceTypeByName) Len() int {
	return len(t)
}

func (t InstanceTypeByName) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t InstanceTypeByName) Less(i, j int) bool {

	// Put General Purpose instance types first
	if strings.HasPrefix(t[i].Name, "General Purpose") {
		if !strings.HasPrefix(t[j].Name, "General Purpose") {
			return true
		}
	} else if strings.HasPrefix(t[j].Name, "General Purpose") {
		return false
	}

	return sortorder.NaturalLess(t[i].Name, t[j].Name)
}

type RegionByName []portaineree.Pair

func (t RegionByName) Len() int {
	return len(t)
}

func (t RegionByName) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t RegionByName) Less(i, j int) bool {
	// Put US Regions at the top.  Similar to AWS Console
	if strings.HasPrefix(t[i].Name, "US ") {
		if !strings.HasPrefix(t[j].Name, "US ") {
			return true
		}
	} else if strings.HasPrefix(t[j].Name, "US ") {
		return false
	}

	return sortorder.NaturalLess(t[i].Name, t[j].Name)
}

type KubeByVersion []portaineree.Pair

func (e KubeByVersion) Len() int {
	return len(e)
}

func (e KubeByVersion) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e KubeByVersion) Less(i, j int) bool {
	return sortorder.NaturalLess(strings.ToLower(e[i].Name), strings.ToLower(e[j].Name))
}
