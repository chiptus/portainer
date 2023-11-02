package endpointrelation

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "endpoint_relations"

// Service represents a service for managing environment(endpoint) relation data.
type Service struct {
	connection      portainer.Connection
	updateStackFn   func(ID portainer.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error
	updateStackFnTx func(tx portainer.Transaction, ID portainer.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error
}

func (service *Service) BucketName() string {
	return BucketName
}

func (service *Service) RegisterUpdateStackFunction(
	updateFunc func(portainer.EdgeStackID, func(*portaineree.EdgeStack)) error,
	updateFuncTx func(portainer.Transaction, portainer.EdgeStackID, func(*portaineree.EdgeStack)) error,
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
func (service *Service) EndpointRelations() ([]portainer.EndpointRelation, error) {
	var all = make([]portainer.EndpointRelation, 0)

	return all, service.connection.GetAll(
		BucketName,
		&portainer.EndpointRelation{},
		dataservices.AppendFn(&all),
	)
}

// EndpointRelation returns a Environment(Endpoint) relation object by EndpointID
func (service *Service) EndpointRelation(endpointID portainer.EndpointID) (*portainer.EndpointRelation, error) {
	var endpointRelation portainer.EndpointRelation
	identifier := service.connection.ConvertToKey(int(endpointID))

	err := service.connection.GetObject(BucketName, identifier, &endpointRelation)
	if err != nil {
		return nil, err
	}

	return &endpointRelation, nil
}

// CreateEndpointRelation saves endpointRelation
func (service *Service) Create(endpointRelation *portainer.EndpointRelation) error {
	err := service.connection.CreateObjectWithId(BucketName, int(endpointRelation.EndpointID), endpointRelation)
	cache.Del(endpointRelation.EndpointID)

	return err
}

// UpdateEndpointRelation updates an Environment(Endpoint) relation object
func (service *Service) UpdateEndpointRelation(endpointID portainer.EndpointID, endpointRelation *portainer.EndpointRelation) error {
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
func (service *Service) DeleteEndpointRelation(endpointID portainer.EndpointID) error {
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

func (service *Service) updateEdgeStacksAfterRelationChange(previousRelationState *portainer.EndpointRelation, updatedRelationState *portainer.EndpointRelation) {
	relations, _ := service.EndpointRelations()

	stacksToUpdate := map[portainer.EdgeStackID]bool{}

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
