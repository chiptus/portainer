package endpointrelation

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "endpoint_relations"

// Service represents a service for managing environment(endpoint) relation data.
type Service struct {
	connection      portainer.Connection
	updateStackFn   func(ID portaineree.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error
	updateStackFnTx func(tx portainer.Transaction, ID portaineree.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error
}

func (service *Service) BucketName() string {
	return BucketName
}

func (service *Service) RegisterUpdateStackFunction(
	updateFunc func(portaineree.EdgeStackID, func(*portaineree.EdgeStack)) error,
	updateFuncTx func(portainer.Transaction, portaineree.EdgeStackID, func(*portaineree.EdgeStack)) error,
) {
	service.updateStackFn = updateFunc
	service.updateStackFnTx = updateFuncTx
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
}

// EndpointRelations returns an array of all EndpointRelations
func (service *Service) EndpointRelations() ([]portaineree.EndpointRelation, error) {
	var all = make([]portaineree.EndpointRelation, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.EndpointRelation{},
		func(obj interface{}) (interface{}, error) {
			r, ok := obj.(*portaineree.EndpointRelation)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to EndpointRelation object")
				return nil, fmt.Errorf("Failed to convert to EndpointRelation object: %s", obj)
			}

			all = append(all, *r)

			return &portaineree.EndpointRelation{}, nil
		})

	return all, err
}

// EndpointRelation returns a Environment(Endpoint) relation object by EndpointID
func (service *Service) EndpointRelation(endpointID portaineree.EndpointID) (*portaineree.EndpointRelation, error) {
	var endpointRelation portaineree.EndpointRelation
	identifier := service.connection.ConvertToKey(int(endpointID))

	err := service.connection.GetObject(BucketName, identifier, &endpointRelation)
	if err != nil {
		return nil, err
	}

	return &endpointRelation, nil
}

// CreateEndpointRelation saves endpointRelation
func (service *Service) Create(endpointRelation *portaineree.EndpointRelation) error {
	err := service.connection.CreateObjectWithId(BucketName, int(endpointRelation.EndpointID), endpointRelation)
	cache.Del(endpointRelation.EndpointID)

	return err
}

// UpdateEndpointRelation updates an Environment(Endpoint) relation object
func (service *Service) UpdateEndpointRelation(endpointID portaineree.EndpointID, endpointRelation *portaineree.EndpointRelation) error {
	previousRelationState, _ := service.EndpointRelation(endpointID)

	identifier := service.connection.ConvertToKey(int(endpointID))
	err := service.connection.UpdateObject(BucketName, identifier, endpointRelation)
	cache.Del(endpointID)
	if err != nil {
		return err
	}

	updatedRelationState, _ := service.EndpointRelation(endpointID)

	service.updateEdgeStacksAfterRelationChange(previousRelationState, updatedRelationState)

	return nil
}

// DeleteEndpointRelation deletes an Environment(Endpoint) relation object
func (service *Service) DeleteEndpointRelation(endpointID portaineree.EndpointID) error {
	deletedRelation, _ := service.EndpointRelation(endpointID)

	identifier := service.connection.ConvertToKey(int(endpointID))
	err := service.connection.DeleteObject(BucketName, identifier)
	cache.Del(endpointID)
	if err != nil {
		return err
	}

	service.updateEdgeStacksAfterRelationChange(deletedRelation, nil)

	return nil
}

func (service *Service) InvalidateEdgeCacheForEdgeStack(edgeStackID portaineree.EdgeStackID) {
	rels, err := service.EndpointRelations()
	if err != nil {
		log.Error().Err(err).Msg("cannot retrieve endpoint relations")
		return
	}

	for _, rel := range rels {
		for id := range rel.EdgeStacks {
			if edgeStackID == id {
				cache.Del(rel.EndpointID)
			}
		}
	}
}

func (service *Service) updateEdgeStacksAfterRelationChange(previousRelationState *portaineree.EndpointRelation, updatedRelationState *portaineree.EndpointRelation) {
	relations, _ := service.EndpointRelations()

	stacksToUpdate := map[portaineree.EdgeStackID]bool{}

	if previousRelationState != nil {
		for stackId, enabled := range previousRelationState.EdgeStacks {
			// flag stack for update if stack is not in the updated relation state
			// = stack has been removed for this relation
			// or this relation has been deleted
			if enabled && (updatedRelationState == nil || !updatedRelationState.EdgeStacks[stackId]) {
				stacksToUpdate[stackId] = true
			}
		}
	}

	if updatedRelationState != nil {
		for stackId, enabled := range updatedRelationState.EdgeStacks {
			// flag stack for update if stack is not in the previous relation state
			// = stack has been added for this relation
			if enabled && (previousRelationState == nil || !previousRelationState.EdgeStacks[stackId]) {
				stacksToUpdate[stackId] = true
			}
		}
	}

	// for each stack referenced by the updated relation
	// list how many time this stack is referenced in all relations
	// in order to update the stack deployments count
	for refStackId, refStackEnabled := range stacksToUpdate {
		if refStackEnabled {
			numDeployments := 0
			for _, r := range relations {
				for sId, enabled := range r.EdgeStacks {
					if enabled && sId == refStackId {
						numDeployments += 1
					}
				}
			}

			service.updateStackFn(refStackId, func(edgeStack *portaineree.EdgeStack) {
				edgeStack.NumDeployments = numDeployments
			})
		}
	}
}
