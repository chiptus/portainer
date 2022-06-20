package podsecurity

import (
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"
	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "podsecurity"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// PodSecurityRule returns a PodSecurityRule object by ID.
func (service *Service) PodSecurityRule(ID podsecurity.PodSecurityRuleID) (*podsecurity.PodSecurityRule, error) {
	var podsecurity podsecurity.PodSecurityRule
	identifier := service.connection.ConvertToKey(int(ID))
	err := service.connection.GetObject(BucketName, identifier, &podsecurity)
	if err != nil {
		return nil, err
	}

	return &podsecurity, nil
}

// PodSecurity returns the first policy with the endpoint id
func (service *Service) PodSecurityByEndpointID(endpointID int) (*podsecurity.PodSecurityRule, error) {
	identifier := service.connection.ConvertToKey(endpointID)
	var rule podsecurity.PodSecurityRule
	err := service.connection.GetObject(BucketName, identifier, &rule)
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

// GetNextIdentifier returns the next identifier for a PodSecurityRule.
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}

// Create creates a new podsecurity.
func (service *Service) Create(podsecurity *podsecurity.PodSecurityRule) error {
	//use Endpoint id as the key
	id := podsecurity.EndpointID
	return service.connection.CreateObjectWithId(BucketName, id, podsecurity)
}

// UpdatePodSecurityRule updates a podsecurity.
func (service *Service) UpdatePodSecurityRule(ID int, podsecurity *podsecurity.PodSecurityRule) error {
	identifier := service.connection.ConvertToKey(ID)
	return service.connection.UpdateObject(BucketName, identifier, podsecurity)
}

// DeletePodSecurityRule deletes a podsecurity.
func (service *Service) DeletePodSecurityRule(ID int) error {
	identifier := service.connection.ConvertToKey(ID)
	return service.connection.DeleteObject(BucketName, identifier)
}
