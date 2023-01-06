package endpoint

import (
	"fmt"
	"sync"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "endpoints"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection portainer.Connection
	mu         sync.RWMutex
	idxEdgeID  map[string]portaineree.EndpointID
	heartbeats sync.Map
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	s := &Service{
		connection: connection,
		idxEdgeID:  make(map[string]portaineree.EndpointID),
	}

	es, err := s.Endpoints()
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		if len(e.EdgeID) > 0 {
			s.idxEdgeID[e.EdgeID] = e.ID
		}

		s.heartbeats.Store(e.ID, e.LastCheckInDate)
	}

	return s, nil
}

// Endpoint returns an environment(endpoint) by ID.
func (service *Service) Endpoint(ID portaineree.EndpointID) (*portaineree.Endpoint, error) {
	var endpoint portaineree.Endpoint
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &endpoint)
	if err != nil {
		return nil, err
	}

	endpoint.LastCheckInDate, _ = service.Heartbeat(ID)

	return &endpoint, nil
}

// UpdateEndpoint updates an environment(endpoint).
func (service *Service) UpdateEndpoint(ID portaineree.EndpointID, endpoint *portaineree.Endpoint) error {
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.UpdateObject(BucketName, identifier, endpoint)
	if err != nil {
		return err
	}

	service.mu.Lock()
	if len(endpoint.EdgeID) > 0 {
		service.idxEdgeID[endpoint.EdgeID] = ID
	}
	service.heartbeats.Store(ID, endpoint.LastCheckInDate)
	service.mu.Unlock()

	cache.Del(endpoint.ID)

	return nil
}

// DeleteEndpoint deletes an environment(endpoint).
func (service *Service) DeleteEndpoint(ID portaineree.EndpointID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	err := service.connection.DeleteObject(BucketName, identifier)
	if err != nil {
		return err
	}

	service.mu.Lock()
	for edgeID, endpointID := range service.idxEdgeID {
		if endpointID == ID {
			delete(service.idxEdgeID, edgeID)
			break
		}
	}
	service.heartbeats.Delete(ID)
	service.mu.Unlock()

	cache.Del(ID)

	return nil
}

// Endpoints return an array containing all the environments(endpoints).
func (service *Service) Endpoints() ([]portaineree.Endpoint, error) {
	var endpoints = make([]portaineree.Endpoint, 0)

	err := service.connection.GetAllWithJsoniter(
		BucketName,
		&portaineree.Endpoint{},
		func(obj interface{}) (interface{}, error) {
			endpoint, ok := obj.(*portaineree.Endpoint)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to Endpoint object")
				return nil, fmt.Errorf("failed to convert to Endpoint object: %s", obj)
			}

			endpoints = append(endpoints, *endpoint)

			return &portaineree.Endpoint{}, nil
		})

	if err != nil {
		return endpoints, err
	}

	for i, e := range endpoints {
		t, _ := service.Heartbeat(e.ID)
		endpoints[i].LastCheckInDate = t
	}

	return endpoints, nil
}

// EndpointIDByEdgeID returns the EndpointID from the given EdgeID using an in-memory index
func (service *Service) EndpointIDByEdgeID(edgeID string) (portaineree.EndpointID, bool) {
	service.mu.RLock()
	endpointID, ok := service.idxEdgeID[edgeID]
	service.mu.RUnlock()

	return endpointID, ok
}

func (service *Service) Heartbeat(endpointID portaineree.EndpointID) (int64, bool) {
	if t, ok := service.heartbeats.Load(endpointID); ok {
		return t.(int64), true
	}

	return 0, false
}

func (service *Service) UpdateHeartbeat(endpointID portaineree.EndpointID) {
	service.heartbeats.Store(endpointID, time.Now().Unix())
}

// CreateEndpoint assign an ID to a new environment(endpoint) and saves it.
func (service *Service) Create(endpoint *portaineree.Endpoint) error {
	err := service.connection.CreateObjectWithId(BucketName, int(endpoint.ID), endpoint)
	if err != nil {
		return err
	}

	service.mu.Lock()
	if len(endpoint.EdgeID) > 0 {
		service.idxEdgeID[endpoint.EdgeID] = endpoint.ID
	}
	service.heartbeats.Store(endpoint.ID, endpoint.LastCheckInDate)
	service.mu.Unlock()

	return nil
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
