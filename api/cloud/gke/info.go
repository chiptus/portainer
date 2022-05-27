package gke

import (
	"context"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	portaineree "github.com/portainer/portainer-ee/api"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type Info struct {
	// Zones are referring to GKE "zones" since we're using zonal provisioning
	// at the moment. We call them "regions" in the response for now so the FE
	// can treat GKE similarly to the other providers.
	Zones []portaineree.Pair `json:"regions"`

	// NodeSizes is a small selected list of GKE Machine Types for users that
	// do not want to use "custom".
	NodeSizes []portaineree.Pair `json:"nodeSizes"`

	// RAM is in Gigabytes with 0.25 increments.
	RAM Spec `json:"ram"`

	// CPU is a count of CPU cores. Must be whole numbers.
	CPU Spec `json:"cpu"`

	// HDD is in Gigabytes and must be whole numbers.
	HDD Spec `json:"hdd"`

	// KubernetesVersions is a list of valid stable kubernetes versions for GKE.
	KubernetesVersions []portaineree.Pair `json:"kubernetesVersions"`

	// Networks is a list of individual networks, with an ID, Name, and Region.
	Networks []Network `json:"networks"`
}

type Spec struct {
	Default float64 `json:"default"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

type Network struct {
	// Region is referring to a GKE Region NOT A ZONE. This means to match the
	// list of zones you must remove the suffix from the zone.
	// e.g us-west1-a -> us-west1
	Region string `json:"region"`

	// Subnets are part of a GKENetwork, but the names are globally unique.
	// This means other networks cannot have subnets with the same names. As a
	// result the frontend only needs to display subnet names rather than
	// Network/Subnet.
	Subnets []SubnetDetails `json:"networks"`
}

type SubnetDetails struct {
	// Network is not exported because FE only needs to list the subnets. We do
	// however want this field internally so we can lookup the Network name
	// tied to a subnet in the provisioning request.
	Network string `json:"-"`

	ID   string `json:"id"`
	Name string `json:"name"`
}

// FetchZones gets a list of "Zones" instead of regions.
//
// In GKE a zone refers to a specific individual datacenter where a node (or
// whole cluster) will exist. A region is a grouping of these zones. Basically
// region = us-east, zone = us-east-a; so we want zones as that's where we're
// actually provisioning the clusters.
func (k Key) FetchZones(ctx context.Context) ([]portaineree.Pair, error) {
	zoneClient, err := compute.NewZonesRESTClient(
		ctx,
		option.WithCredentialsJSON(k.Bytes),
	)
	if err != nil {
		return nil, fmt.Errorf("GKE key may be invalid: %w", err)
	}
	defer zoneClient.Close()

	zoneReq := &computepb.ListZonesRequest{
		Project: k.ProjectID,
	}

	regions := make([]portaineree.Pair, 0)
	zoneIt := zoneClient.List(ctx, zoneReq)
	for {
		resp, err := zoneIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		id := resp.GetName()
		prettyName := strings.ReplaceAll(id, "-", " ")
		prettyName = strings.Title(prettyName)
		prettyName = strings.ReplaceAll(prettyName, "Us", "US")
		prettyName = strings.ReplaceAll(prettyName, "america", " America")

		r := portaineree.Pair{
			Name:  prettyName,
			Value: id,
		}
		regions = append(regions, r)
	}
	return regions, err
}

// FetchNetworks gets a list of networks which can be used to provision a
// GKE cluster.
func (k Key) FetchNetworks(ctx context.Context) ([]Network, error) {
	nets := make([]Network, 0)
	networkClient, err := compute.NewNetworksRESTClient(
		ctx,
		option.WithCredentialsJSON(k.Bytes),
	)
	if err != nil {
		return nil, fmt.Errorf("GKE key may be invalid: %w", err)
	}
	defer networkClient.Close()

	networkReq := &computepb.ListNetworksRequest{
		Project: k.ProjectID,
	}

	// GKE returns the subnets as a list of URLs. Unfortunately this means if
	// we the actual names we have to parse these URLs.
	splitSlash := func(c rune) bool {
		return c == '/'
	}

	type subInfo struct {
		name   string
		subnet string
	}
	// netsMap maps net (containing a name and region) to a list of subnets
	// which are part of that region. This is a temporary intermediate step
	// before we write this into the GKENetwork struct as a slice.
	netsMap := make(map[string][]subInfo, 0)
	networkIt := networkClient.List(ctx, networkReq)
	for {
		resp, err := networkIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		name := resp.GetName()

		respSubnets := resp.GetSubnetworks()
		for _, respSubnet := range respSubnets {
			sliceNetURL := strings.FieldsFunc(respSubnet, splitSlash)

			if len(sliceNetURL) < 10 {
				return nil, fmt.Errorf("failed while reading GKE's network list: could not parse %v", sliceNetURL)
			}
			region := sliceNetURL[7]
			subnet := sliceNetURL[9]
			sub := subInfo{name: name, subnet: subnet}
			netsMap[region] = append(netsMap[region], sub)
		}
	}

	for region, networks := range netsMap {
		details := make([]SubnetDetails, 0)
		for _, info := range networks {
			detail := SubnetDetails{
				Network: info.name,
				ID:      info.subnet,
				Name:    info.subnet,
			}
			details = append(details, detail)
		}
		net := Network{
			Region:  region,
			Subnets: details,
		}
		nets = append(nets, net)
	}
	return nets, err
}

// FetchMachines gets a list of machine types which can be used to provision
// a GKE cluster instead of a custom type.
func (k Key) FetchMachines(ctx context.Context, zone string) ([]portaineree.Pair, error) {
	machines := make([]portaineree.Pair, 0)

	zoneClient, err := compute.NewMachineTypesRESTClient(
		ctx,
		option.WithCredentialsJSON(k.Bytes),
	)
	if err != nil {
		return nil, fmt.Errorf("GKE key may be invalid: %w", err)
	}
	defer zoneClient.Close()

	filter := `( zone = ` + zone + ` ) ( name : "e2*" )`
	machineReq := &computepb.AggregatedListMachineTypesRequest{
		Project: k.ProjectID,
		Filter:  &filter,
	}

	machineIt := zoneClient.AggregatedList(ctx, machineReq)
	for {
		resp, err := machineIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, mt := range resp.Value.MachineTypes {
			var machine portaineree.Pair

			desc := strings.Replace(*mt.Description, "Efficient Instance, ", "", 1)

			machine.Name = *mt.Name + " (" + desc + ")"
			machine.Value = *mt.Name
			machines = append(machines, machine)
		}
	}

	// Add default "custom" type.
	machines = append(machines, portaineree.Pair{
		Name:  "custom",
		Value: "custom",
	})

	return machines, nil
}

// FetchVersions gets a list of kubernetes versions which can be used to
// provision a GKE cluster.
//
// The list of kubernetes versions is region specific for GKE. However, they
// appear to pretty much always be the same. Rather than making our API much
// more complex we've decided for now we'll simply grab the version for
// whichever region is listed first.
//
// Additionally, for portainer we only fetch versions which are part of the
// STABLE channel.
func (k Key) FetchVersions(ctx context.Context, zone string) ([]portaineree.Pair, error) {
	var versions []string
	containerService, err := container.NewService(
		ctx,
		option.WithCredentialsJSON(k.Bytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating container service: %v", err)
	}
	zoneService := container.NewProjectsZonesService(containerService)

	config, err := zoneService.GetServerconfig(k.ProjectID, zone).Do()
	if err != nil {
		return nil, fmt.Errorf("failed fetching list of kubernetes versions")
	}

	channels := config.Channels
	for _, channel := range channels {
		if channel.Channel == "STABLE" {
			versions = channel.ValidVersions
		}

	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("failed parsing list of kubernetes versions")
	}

	pairs := make([]portaineree.Pair, 0)
	for _, v := range versions {
		pair := portaineree.Pair{
			Name:  v,
			Value: v,
		}
		pairs = append(pairs, pair)
	}

	return pairs, err
}
