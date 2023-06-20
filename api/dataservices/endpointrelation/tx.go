package endpointrelation

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

// EndpointRelations returns an array of all EndpointRelations
func (service ServiceTx) EndpointRelations() ([]portaineree.EndpointRelation, error) {
	var all = make([]portaineree.EndpointRelation, 0)

	return all, service.tx.GetAll(
		BucketName,
		&portaineree.EndpointRelation{},
		dataservices.AppendFn(&all),
	)
}

// EndpointRelation returns a Environment(Endpoint) relation object by EndpointID
func (service ServiceTx) EndpointRelation(endpointID portaineree.EndpointID) (*portaineree.EndpointRelation, error) {
	var endpointRelation portaineree.EndpointRelation
	identifier := service.service.connection.ConvertToKey(int(endpointID))

	err := service.tx.GetObject(BucketName, identifier, &endpointRelation)
	if err != nil {
		return nil, err
	}

	return &endpointRelation, nil
}

// CreateEndpointRelation saves endpointRelation
func (service ServiceTx) Create(endpointRelation *portaineree.EndpointRelation) error {
	err := service.tx.CreateObjectWithId(BucketName, int(endpointRelation.EndpointID), endpointRelation)
	cache.Del(endpointRelation.EndpointID)

	return err
}

// UpdateEndpointRelation updates an Environment(Endpoint) relation object
func (service ServiceTx) UpdateEndpointRelation(endpointID portaineree.EndpointID, endpointRelation *portaineree.EndpointRelation) error {
	previousRelationState, _ := service.EndpointRelation(endpointID)

	identifier := service.service.connection.ConvertToKey(int(endpointID))
	err := service.tx.UpdateObject(BucketName, identifier, endpointRelation)
	cache.Del(endpointID)
	if err != nil {
		return err
	}

	updatedRelationState, _ := service.EndpointRelation(endpointID)

	service.updateEdgeStacksAfterRelationChange(previousRelationState, updatedRelationState)

	return nil
}

// DeleteEndpointRelation deletes an Environment(Endpoint) relation object
func (service ServiceTx) DeleteEndpointRelation(endpointID portaineree.EndpointID) error {
	deletedRelation, _ := service.EndpointRelation(endpointID)

	identifier := service.service.connection.ConvertToKey(int(endpointID))
	err := service.tx.DeleteObject(BucketName, identifier)
	cache.Del(endpointID)
	if err != nil {
		return err
	}

	service.updateEdgeStacksAfterRelationChange(deletedRelation, nil)

	return nil
}

func (service ServiceTx) InvalidateEdgeCacheForEdgeStack(edgeStackID portaineree.EdgeStackID) {
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

func (service ServiceTx) updateEdgeStacksAfterRelationChange(previousRelationState *portaineree.EndpointRelation, updatedRelationState *portaineree.EndpointRelation) {
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

			service.service.updateStackFnTx(service.tx, refStackId, func(edgeStack *portaineree.EdgeStack) {
				edgeStack.NumDeployments = numDeployments
			})
		}
	}
}
