package providers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/types"
	"github.com/portainer/portainer-ee/api/internal/iprange"
)

type (
	Microk8sTestSSHPayload struct {
		Nodes        []string                 `validate:"required" json:"nodeIPs"`
		CredentialID models.CloudCredentialID `validate:"required" json:"credentialID" example:"1"`
	}

	Microk8sProvisionPayload struct {
		MasterNodes       []string `validate:"required" json:"masterNodes"`
		WorkerNodes       []string `json:"workerNodes"`
		KubernetesVersion string   `validate:"required" json:"kubernetesVersion"`
		Addons            []string `json:"addons"`

		DefaultProvisionPayload
	}

	Microk8sScaleClusterPayload struct {
		MasterNodesToAdd []string `json:"masterNodesToAdd"`
		WorkerNodesToAdd []string `json:"workerNodesToAdd"`
		NodesToRemove    []string `json:"nodesToRemove"`
	}

	Microk8sUpdateAddonsPayload struct {
		Addons []string `json:"addons"`
	}
)

func (payload *Microk8sTestSSHPayload) Validate(r *http.Request) error {
	return validateNodes(payload.Nodes)
}

func (payload *Microk8sScaleClusterPayload) Validate(r *http.Request) error {
	if len(payload.MasterNodesToAdd) >= 0 && len(payload.WorkerNodesToAdd) >= 0 {
		nodes := append(payload.MasterNodesToAdd, payload.WorkerNodesToAdd...)
		err := validateNodes(nodes)
		if err != nil {
			return err
		}

		return nil
	}

	if len(payload.NodesToRemove) > 0 {
		err := validateNodes(payload.NodesToRemove)
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("invalid request payload, nodes to add or remove is empty")
}

func validateNodes(nodesOrNodeRanges []string) error {
	// for each ip range, check whether it overlaps any other ip range provided
	var ranges []iprange.IPRange
	var errors []error

	for _, node := range nodesOrNodeRanges {
		// parse the ranges
		r, err := iprange.Parse(node)
		if err != nil {
			if node == "" {
				err = fmt.Errorf("node range cannot be empty")
			} else if govalidator.IsHost(node) {
				// TODO: future - support hostnames by skipping here
				// skip hostnames
				// 	continue
				err = fmt.Errorf("parse %s failed, hostnames are not currently supported", node)
			}

			errors = append(errors, err)
		}

		// add parsed range to the list
		ranges = append(ranges, r)
	}

	// check for range overlaps
	for i, r1 := range ranges {
		for j, r2 := range ranges {
			// skip self
			if i == j {
				continue
			}

			if r1.Overlaps(r2) {
				errors = append(errors, fmt.Errorf("node ranges overlap, %v and %v", r1, r2))
				break
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid node address: %v", errors)
	}

	return nil
}

func (payload *Microk8sProvisionPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("invalid cluster name")
	}

	if govalidator.IsNonPositive(float64(payload.CredentialID)) {
		return errors.New("invalid credentials")
	}

	if len(payload.MasterNodes) == 0 {
		return errors.New("no master nodes specified")
	}

	nodes := append(payload.MasterNodes, payload.WorkerNodes...)
	err := validateNodes(nodes)
	if err != nil {
		return err
	}

	return nil
}

func (payload *Microk8sProvisionPayload) GetCloudProvider(string) (*portaineree.CloudProvider, error) {
	cloudProvider, ok := types.CloudProvidersMap[types.CloudProviderShortName(portaineree.CloudProviderMicrok8s)]
	if !ok {
		return nil, errors.New("invalid cloud provider")
	}

	cloudProvider.CredentialID = payload.CredentialID
	cloudProvider.NodeCount = payload.NodeCount
	if payload.Addons != nil {
		addons := strings.Join(payload.Addons, ", ")
		cloudProvider.Addons = &addons
	}

	nodes := strings.Join(payload.MasterNodes, ",")
	if len(payload.WorkerNodes) > 0 {
		nodes = fmt.Sprintf("%s,%s", nodes, strings.Join(payload.WorkerNodes, ","))
	}
	cloudProvider.NodeIPs = &nodes
	return &cloudProvider, nil
}

func (payload *Microk8sProvisionPayload) GetCloudProvisioningRequest(endpointID portaineree.EndpointID, _ string) *portaineree.CloudProvisioningRequest {

	// nodes have been Parsed before inside Validate so skip the error check
	var masters []string
	for _, node := range payload.MasterNodes {
		r, _ := iprange.Parse(node)
		masters = append(masters, r.Expand()...)
	}

	var workers []string
	for _, node := range payload.WorkerNodes {
		r, _ := iprange.Parse(node)
		workers = append(workers, r.Expand()...)
	}

	request := &portaineree.CloudProvisioningRequest{
		EndpointID:            endpointID,
		Provider:              portaineree.CloudProviderMicrok8s,
		Name:                  payload.Name,
		CredentialID:          payload.CredentialID,
		NodeCount:             payload.NodeCount,
		MasterNodes:           masters,
		WorkerNodes:           workers,
		Addons:                payload.Addons,
		CustomTemplateID:      payload.Meta.CustomTemplateID,
		CustomTemplateContent: payload.Meta.CustomTemplateContent,
		KubernetesVersion:     payload.KubernetesVersion,
	}

	return request
}

func (payload *Microk8sUpdateAddonsPayload) Validate(r *http.Request) error {
	return nil
}
