package microk8s

import (
	"encoding/json"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/rs/zerolog/log"
)

type (
	Microk8sScalingRequest struct {
		EndpointID portaineree.EndpointID `json:"EndpointID"`

		// scaling up
		MasterNodesToAdd []string `json:"MasterNodesToAdd,omitempty"`
		WorkerNodesToAdd []string `json:"WorkerNodesToAdd,omitempty"`

		// scaling down or removing nodes
		NodesToRemove []string `json:"NodesToRemove,omitempty"`
	}

	Microk8sScalingRequestFactory struct {
		scalingRequest *Microk8sScalingRequest
		dataStore      dataservices.DataStore
		setMessage     func(summary, detail, operatingStatus string) error
		clientFactory  *kubecli.ClientFactory
	}
)

func (r *Microk8sScalingRequest) Provider() string {
	return portaineree.CloudProviderMicrok8s
}

func (r *Microk8sScalingRequest) String() string {
	// convert to json
	b, err := json.Marshal(*r)
	if err != nil {
		return ""
	}
	return string(b)
}

func NewMicrok8sScalingRequestFactory(
	dataStore dataservices.DataStore,
	clientFactory *kubecli.ClientFactory,
	scalingRequest *Microk8sScalingRequest,
	setMessage func(summary, detail, operatingStatus string) error,
) *Microk8sScalingRequestFactory {
	return &Microk8sScalingRequestFactory{
		dataStore:      dataStore,
		clientFactory:  clientFactory,
		scalingRequest: scalingRequest,
		setMessage:     setMessage,
	}
}

func (s *Microk8sScalingRequestFactory) Process() error {
	log.Debug().Msgf("Processing microk8s scaling request for environment %d", s.scalingRequest.EndpointID)

	s.setMessage("Scaling cluster", "Scaling in progress", "processing")

	endpoint, err := s.dataStore.Endpoint().Endpoint(s.scalingRequest.EndpointID)
	if err != nil {
		details := fmt.Sprintf("Scaling error: %v", err)
		s.setMessage("Failed to scale cluster", details, "warning")
		return fmt.Errorf("failed to retrieve environment %d. %w", s.scalingRequest.EndpointID, err)
	}

	if endpoint.CloudProvider == nil {
		return fmt.Errorf("environment %d was not provisioned from Portainer", s.scalingRequest.EndpointID)
	}

	credentials, err := s.dataStore.CloudCredential().Read(endpoint.CloudProvider.CredentialID)
	if err != nil {
		details := fmt.Sprintf("Scaling error: %v", err)
		s.setMessage("Failed to scale cluster", details, "warning")
		return fmt.Errorf("failed to retrieve credentials for endpoint %d. %w", s.scalingRequest.EndpointID, err)
	}

	mk8sCluster := Microk8sCluster{
		DataStore:         s.dataStore,
		KubeClientFactory: s.clientFactory,
		SetMessage:        s.setMessage,
	}
	if len(s.scalingRequest.MasterNodesToAdd) > 0 || len(s.scalingRequest.WorkerNodesToAdd) > 0 {
		s.setMessage("Scaling cluster", "Scaling up in progress", "processing")
		err = mk8sCluster.AddNodes(endpoint, credentials, s.scalingRequest)
	} else if len(s.scalingRequest.NodesToRemove) > 0 {
		s.setMessage("Scaling cluster", "Scaling down in progress", "processing")
		err = mk8sCluster.RemoveNodes(endpoint, credentials, s.scalingRequest)
	}

	if err != nil {
		details := fmt.Sprintf("Scaling error: %v", err)
		s.setMessage("Failed to scale cluster", details, "warning")
		return err
	}

	s.setMessage("Scaling up cluster", "Scaling complete", "")
	return nil
}
