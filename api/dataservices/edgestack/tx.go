package edgestack

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
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

// EdgeStacks returns an array containing all edge stacks
func (service ServiceTx) EdgeStacks() ([]portaineree.EdgeStack, error) {
	var stacks = make([]portaineree.EdgeStack, 0)

	err := service.tx.GetAll(
		BucketName,
		&portaineree.EdgeStack{},
		func(obj interface{}) (interface{}, error) {
			stack, ok := obj.(*portaineree.EdgeStack)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to EdgeStack object")
				return nil, fmt.Errorf("failed to convert to EdgeStack object: %s", obj)
			}

			stacks = append(stacks, *stack)

			return &portaineree.EdgeStack{}, nil
		})

	return stacks, err
}

// EdgeStack returns an Edge stack by ID.
func (service ServiceTx) EdgeStack(ID portainer.EdgeStackID) (*portaineree.EdgeStack, error) {
	var stack portaineree.EdgeStack
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &stack)
	if err != nil {
		return nil, err
	}

	return &stack, nil
}

// EdgeStackVersion returns the version of the given edge stack ID directly from an in-memory index
func (service ServiceTx) EdgeStackVersion(ID portainer.EdgeStackID) (int, bool) {
	service.service.mu.RLock()
	v, ok := service.service.idxVersion[ID]
	service.service.mu.RUnlock()

	return v, ok
}

// CreateEdgeStack saves an Edge stack object to db.
func (service ServiceTx) Create(id portainer.EdgeStackID, edgeStack *portaineree.EdgeStack) error {
	edgeStack.ID = id

	err := service.tx.CreateObjectWithId(
		BucketName,
		int(edgeStack.ID),
		edgeStack,
	)
	if err != nil {
		return err
	}

	service.service.mu.Lock()
	service.service.idxVersion[id] = edgeStack.Version
	service.service.cacheInvalidationFn(service.tx, id)
	service.service.mu.Unlock()

	return nil
}

// UpdateEdgeStack updates an Edge stack.
func (service ServiceTx) UpdateEdgeStack(ID portainer.EdgeStackID, edgeStack *portaineree.EdgeStack, cleanupCache bool) error {
	service.service.mu.Lock()
	defer service.service.mu.Unlock()

	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.UpdateObject(BucketName, identifier, edgeStack)
	if err != nil {
		return err
	}

	service.service.idxVersion[ID] = edgeStack.Version

	if cleanupCache {
		service.service.cacheInvalidationFn(service.tx, ID)
	}

	return nil
}

// Deprecated: use UpdateEdgeStack inside a transaction instead.
func (service ServiceTx) UpdateEdgeStackFunc(ID portainer.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error {
	edgeStack, err := service.EdgeStack(ID)
	if err != nil {
		return err
	}

	updateFunc(edgeStack)

	return service.UpdateEdgeStack(ID, edgeStack, true)
}

// DeleteEdgeStack deletes an Edge stack.
func (service ServiceTx) DeleteEdgeStack(ID portainer.EdgeStackID) error {
	service.service.mu.Lock()
	defer service.service.mu.Unlock()

	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.DeleteObject(BucketName, identifier)
	if err != nil {
		return err
	}

	delete(service.service.idxVersion, ID)

	service.service.cacheInvalidationFn(service.tx, ID)

	return nil
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service ServiceTx) GetNextIdentifier() int {
	return service.tx.GetNextIdentifier(BucketName)
}
