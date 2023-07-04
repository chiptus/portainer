package endpoint

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

// Endpoint returns an environment(endpoint) by ID.
func (service ServiceTx) Endpoint(ID portaineree.EndpointID) (*portaineree.Endpoint, error) {
	var endpoint portaineree.Endpoint
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &endpoint)
	if err != nil {
		return nil, err
	}

	return &endpoint, nil
}

func (service ServiceTx) SetMessage(ID portaineree.EndpointID, statusMessage portaineree.EndpointStatusMessage) error {
	endpoint, err := service.Endpoint(ID)
	if err != nil {
		return err
	}

	endpoint.StatusMessage = statusMessage

	return service.UpdateEndpoint(ID, endpoint)
}

// UpdateEndpoint updates an environment(endpoint).
func (service ServiceTx) UpdateEndpoint(ID portaineree.EndpointID, endpoint *portaineree.Endpoint) error {
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.UpdateObject(BucketName, identifier, endpoint)
	if err != nil {
		return err
	}

	service.service.mu.Lock()
	if len(endpoint.EdgeID) > 0 {
		service.service.idxEdgeID[endpoint.EdgeID] = ID
	}
	service.service.heartbeats.Store(ID, endpoint.LastCheckInDate)
	service.service.mu.Unlock()

	cache.Del(endpoint.ID)

	return nil
}

// DeleteEndpoint deletes an environment(endpoint).
func (service ServiceTx) DeleteEndpoint(ID portaineree.EndpointID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.DeleteObject(BucketName, identifier)
	if err != nil {
		return err
	}

	service.service.mu.Lock()
	for edgeID, endpointID := range service.service.idxEdgeID {
		if endpointID == ID {
			delete(service.service.idxEdgeID, edgeID)
			break
		}
	}
	service.service.heartbeats.Delete(ID)
	service.service.mu.Unlock()

	cache.Del(ID)

	return nil
}

// Endpoints return an array containing all the environments(endpoints).
func (service ServiceTx) Endpoints() ([]portaineree.Endpoint, error) {
	var endpoints = make([]portaineree.Endpoint, 0)

	return endpoints, service.tx.GetAllWithJsoniter(
		BucketName,
		&portaineree.Endpoint{},
		dataservices.AppendFn(&endpoints),
	)
}

func (service ServiceTx) EndpointIDByEdgeID(edgeID string) (portaineree.EndpointID, bool) {
	log.Error().Str("func", "EndpointIDByEdgeID").Msg("cannot be called inside a transaction")

	return 0, false
}

func (service ServiceTx) Heartbeat(endpointID portaineree.EndpointID) (int64, bool) {
	log.Error().Str("func", "Heartbeat").Msg("cannot be called inside a transaction")

	return 0, false
}

func (service ServiceTx) UpdateHeartbeat(endpointID portaineree.EndpointID) {
	log.Error().Str("func", "UpdateHeartbeat").Msg("cannot be called inside a transaction")
}

// CreateEndpoint assign an ID to a new environment(endpoint) and saves it.
func (service ServiceTx) Create(endpoint *portaineree.Endpoint) error {
	err := service.tx.CreateObjectWithId(BucketName, int(endpoint.ID), endpoint)
	if err != nil {
		return err
	}

	service.service.mu.Lock()
	if len(endpoint.EdgeID) > 0 {
		service.service.idxEdgeID[endpoint.EdgeID] = endpoint.ID
	}
	service.service.heartbeats.Store(endpoint.ID, endpoint.LastCheckInDate)
	service.service.mu.Unlock()

	return nil
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service ServiceTx) GetNextIdentifier() int {
	return service.tx.GetNextIdentifier(BucketName)
}
