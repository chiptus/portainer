package eks

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

const (
	accelerated      = "Accelerated Computing"
	computeOptimized = "Compute Optimized"
	generalPurpose   = "General Purpose"
	memoryOptimized  = "Memory Optimized"
	storageOptimized = "Storage Optimized"
)

// getInstanceUseCase returns the use case given an aws instance type
// e.g.  m5 returns "General Purpose"
func getInstanceUseCase(instanceType string) string {
	if len(instanceType) < 3 {
		return ""
	}

	// Check prefixes of instance types defined here:
	//   https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html

	useCases := map[string]string{
		"hpc": computeOptimized,
		"dl":  accelerated,
		"in":  accelerated,
		"vt":  accelerated,
		"im":  storageOptimized,
		"is":  storageOptimized,
		"u-":  memoryOptimized,
		"c":   computeOptimized,
		"d":   storageOptimized,
		"h":   storageOptimized,
		"i":   storageOptimized,
		"f":   accelerated,
		"g":   accelerated,
		"p":   accelerated,
		"m":   generalPurpose,
		"t":   generalPurpose,
		"r":   memoryOptimized,
		"x":   memoryOptimized,
		"z":   memoryOptimized,
	}

	// match prefix.  Longest match wins
	for l := 3; l > 0; l-- {
		prefix := instanceType[:l]
		if useCase, ok := useCases[prefix]; ok {
			return useCase
		}
	}

	// blank if not matched
	return ""
}

// formats memory in MB to either MB, GB or TB
func formatMemorySize(sizeInMiB int64) string {
	if sizeInMiB < 1024 {
		return fmt.Sprintf("%dMB", sizeInMiB)
	}
	if sizeInMiB < 1024*1024 {
		return fmt.Sprintf("%dGB", sizeInMiB/1024)
	}

	return fmt.Sprintf("%dTB", sizeInMiB/1024/1024)
}

func getRegions(cfg aws.Config) ([]portaineree.Pair, error) {
	// Get a list of regions from our default region
	svc := ec2.NewFromConfig(cfg)
	result, err := svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	var lock sync.Mutex
	errs, _ := errgroup.WithContext(context.TODO())

	var regions []portaineree.Pair
	for _, r := range result.Regions {
		region := *r.RegionName

		// Get long name from ssm, which is very slow so we parallelise it with goroutines
		errs.Go(func() error {
			longName, err := getRegionLongName(cfg, region)
			if err != nil {
				return errors.Wrapf(err, "Could not get long name for %s", region)
			}

			lock.Lock()
			regions = append(regions, portaineree.Pair{Name: longName, Value: region})
			lock.Unlock()

			return nil
		})
	}

	err = errs.Wait()
	if err != nil {
		// If one region fails don't return an error or it will fail all.
		// At least this way we get something back.
		log.Warnf("Get long region names failed: %v", err)
	}

	sort.Stable(RegionByName(regions))
	return regions, nil
}

func getRegionLongName(cfg aws.Config, shortName string) (string, error) {
	ssmsvc := ssm.NewFromConfig(cfg)

	regionInfo, err := ssmsvc.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name: aws.String("/aws/service/global-infrastructure/regions/" + shortName + "/longName"),
	})
	if err != nil {
		return "", err
	}

	return *regionInfo.Parameter.Value, err
}

func FetchInfo(accessKeyId, secretAccessKey string) (*EksInfo, error) {
	log.Debug("[cloud] [message: sending cloud provider info request] [provider: amazon]")

	cfg, err := loadConfig(accessKeyId, secretAccessKey, nil)
	if err != nil {
		return nil, err
	}

	eksInfo := &EksInfo{}
	eksInfo.Regions, err = getRegions(cfg)
	if err != nil {
		return nil, err
	}

	kubeVersions, err := getKubernetesVersions(cfg)
	if err != nil {
		return nil, err
	}
	eksInfo.KubernetesVersions = append(eksInfo.KubernetesVersions, kubeVersions...)

	eksInfo.AmiTypes = getAmiTypes()

	var lock sync.Mutex
	errs, _ := errgroup.WithContext(context.TODO())

	// For each region, connect to it and query what instance types are available
	eksInfo.InstanceTypes = make(map[string][]InstanceType)

	for _, r := range eksInfo.Regions {
		region := r.Value
		errs.Go(func() error {
			cfg, err := loadConfig(accessKeyId, secretAccessKey, &region)
			if err != nil {
				return err
			}

			svc := ec2.NewFromConfig(cfg)

			// Get instance types for each region.
			it, err := svc.DescribeInstanceTypes(context.TODO(), getInstanceTypesInput())
			if err != nil {
				return errors.Wrapf(err, "Could not get instance types for region %s", region)
			}

			// copy into our info map
			lock.Lock()
			defer lock.Unlock()
			eksInfo.InstanceTypes[region] = createInstanceTypes(it.InstanceTypes)

			return nil
		})
	}

	err = errs.Wait()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get describe instance types for some regions")
	}

	return eksInfo, nil
}

func createInstanceTypes(instanceTypes []ec2types.InstanceTypeInfo) []InstanceType {
	var instanceTypesList []InstanceType

	for _, v := range instanceTypes {
		t := createInstanceType(v)

		// attach compatible ami's to the instance type.
		// taking into account supported architecture and whether the instance type has a GPU or not
		for _, ami := range getAmiTypes() {
			for _, arch := range v.ProcessorInfo.SupportedArchitectures {

				translate := map[ec2types.ArchitectureType]string{
					"x86_64": "x86_64",
					"arm64":  "ARM",
				}

				if strings.Contains(ami.Value, translate[arch]) {
					gpuAmiType := strings.HasSuffix(ami.Value, "GPU") || strings.HasSuffix(ami.Value, "NVIDIA")
					if (v.GpuInfo != nil && !gpuAmiType) || (gpuAmiType && v.GpuInfo == nil) {
						continue
					}

					// don't add duplicates
					if !slices.Contains(t.CompatibleAmiTypes, ami.Value) {
						t.CompatibleAmiTypes = append(t.CompatibleAmiTypes, ami.Value)
					}
				}
			}
		}

		instanceTypesList = append(instanceTypesList, t)
	}

	sort.Stable(InstanceTypeByName(instanceTypesList))

	return instanceTypesList
}

func createInstanceType(v ec2types.InstanceTypeInfo) InstanceType {
	gpuInfo := ""
	if v.GpuInfo != nil {
		gpuInfo = fmt.Sprintf("- %d %s %s GPUs", *v.GpuInfo.Gpus[0].Count, *v.GpuInfo.Gpus[0].Manufacturer, *v.GpuInfo.Gpus[0].Name)
	}

	t := InstanceType{}
	t.Name = fmt.Sprintf("%s: %s (%d vCPU - %s RAM - network %s%s)",
		getInstanceUseCase(string(v.InstanceType)),
		string(v.InstanceType),
		(*v.VCpuInfo.DefaultCores)*(*v.VCpuInfo.DefaultThreadsPerCore),
		formatMemorySize(*v.MemoryInfo.SizeInMiB),
		*v.NetworkInfo.NetworkPerformance,
		gpuInfo)
	t.Value = string(v.InstanceType)

	return t
}

func getAmiTypes() []portaineree.Pair {

	// a predefined list of AMI types, unfortunately I can't find anywhere in the API to read this
	// https://docs.aws.amazon.com/eks/latest/APIReference/API_Nodegroup.html
	return []portaineree.Pair{
		{Name: "Amazon Linux 2 (AL2_x86_64)", Value: "AL2_x86_64"},
		{Name: "Amazon Linux 2 GPU Enabled (AL2_x86_64_GPU)", Value: "AL2_x86_64_GPU"},
		{Name: "Amazon Linux 2 Arm (AL2_ARM_64)", Value: "AL2_ARM_64"},
		{Name: "Bottlerocket (BOTTLEROCKET_x86_64)", Value: "BOTTLEROCKET_x86_64"},
		{Name: "Bottlerocket Arm (BOTTLEROCKET_ARM_64)", Value: "BOTTLEROCKET_ARM_64"},
		{Name: "Bottlerocket NVIDIA (BOTTLEROCKET_x86_64_NVIDIA)", Value: "BOTTLEROCKET_x86_64_NVIDIA"},
		{Name: "Bottlerocket Arm NVIDIA (BOTTLEROCKET_ARM_64_NVIDIA)", Value: "BOTTLEROCKET_ARM_64_NVIDIA"},
	}
}

func getInstanceTypesInput() *ec2.DescribeInstanceTypesInput {
	// filter instances to only the types we want, see below
	return &ec2.DescribeInstanceTypesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("current-generation"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("processor-info.supported-architecture"),
				Values: []string{"x86_64", "arm64"},
			},
			{
				Name:   aws.String("supported-virtualization-type"),
				Values: []string{"hvm"},
			},
		},
	}
}

func getKubernetesVersions(cfg aws.Config) ([]portaineree.Pair, error) {
	// EKS does not provide an api to get kubernetes versions directly
	// This is a hack to get it via the addons api
	// We can look at the addons and versions and check the compatibile versions of kubernetes for these addons

	versions := make(map[string]string, 0)

	addonVersions, err := eks.NewFromConfig(cfg).DescribeAddonVersions(context.TODO(), &eks.DescribeAddonVersionsInput{})
	if err != nil {
		return nil, err
	}

	// Grab all the kube versions (will be several duplicates). Manage it with a map.
	for _, versionInfo := range addonVersions.Addons {
		for _, addonVer := range versionInfo.AddonVersions {
			for _, compatVer := range addonVer.Compatibilities {
				versions[*compatVer.ClusterVersion] = *compatVer.ClusterVersion
			}
		}
	}

	// Now change the map into the format we require
	var kubeVersions []portaineree.Pair
	for v := range versions {
		kubeVersions = append(kubeVersions, portaineree.Pair{Name: v, Value: v})
	}
	sort.Sort(sort.Reverse(KubeByVersion(kubeVersions)))
	kubeVersions = append([]portaineree.Pair{{Name: "latest", Value: ""}}, kubeVersions...)

	return kubeVersions, nil
}
