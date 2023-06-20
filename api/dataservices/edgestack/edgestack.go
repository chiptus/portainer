package edgestack

import (
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "edge_stack"

// Service represents a service for managing Edge stack data.
type Service struct {
	connection          portainer.Connection
	idxVersion          map[portaineree.EdgeStackID]int
	mu                  sync.RWMutex
	cacheInvalidationFn func(portaineree.EdgeStackID)
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection, cacheInvalidationFn func(portaineree.EdgeStackID)) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	s := &Service{
		connection:          connection,
		idxVersion:          make(map[portaineree.EdgeStackID]int),
		cacheInvalidationFn: cacheInvalidationFn,
	}

	if s.cacheInvalidationFn == nil {
		s.cacheInvalidationFn = func(portaineree.EdgeStackID) {}
	}

	es, err := s.EdgeStacks()
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		s.idxVersion[e.ID] = e.Version
	}

	return s, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
}

// EdgeStacks returns an array containing all edge stacks
func (service *Service) EdgeStacks() ([]portaineree.EdgeStack, error) {
	var stacks = make([]portaineree.EdgeStack, 0)

	return stacks, service.connection.GetAll(
		BucketName,
		&portaineree.EdgeStack{},
		dataservices.AppendFn(&stacks),
	)
}

// EdgeStack returns an Edge stack by ID.
func (service *Service) EdgeStack(ID portaineree.EdgeStackID) (*portaineree.EdgeStack, error) {
	var stack portaineree.EdgeStack
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &stack)
	if err != nil {
		return nil, err
	}

	return &stack, nil
}

// EdgeStackVersion returns the version of the given edge stack ID directly from an in-memory index
func (service *Service) EdgeStackVersion(ID portaineree.EdgeStackID) (int, bool) {
	service.mu.RLock()
	v, ok := service.idxVersion[ID]
	service.mu.RUnlock()

	return v, ok
}

// CreateEdgeStack saves an Edge stack object to db.
func (service *Service) Create(id portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error {
	edgeStack.ID = id

	err := service.connection.CreateObjectWithId(
		BucketName,
		int(edgeStack.ID),
		edgeStack,
	)
	if err != nil {
		return err
	}

	service.mu.Lock()
	service.idxVersion[id] = edgeStack.Version
	service.cacheInvalidationFn(id)
	service.mu.Unlock()

	return nil
}

// Deprecated: Use UpdateEdgeStackFunc instead.
func (service *Service) UpdateEdgeStack(ID portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.UpdateObject(BucketName, identifier, edgeStack)
	if err != nil {
		return err
	}

	service.idxVersion[ID] = edgeStack.Version
	service.cacheInvalidationFn(ID)

	return nil
}

// UpdateEdgeStackFunc updates an Edge stack inside a transaction avoiding data races.
func (service *Service) UpdateEdgeStackFunc(ID portaineree.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error {
	id := service.connection.ConvertToKey(int(ID))
	edgeStack := &portaineree.EdgeStack{}

	service.mu.Lock()
	defer service.mu.Unlock()

	return service.connection.UpdateObjectFunc(BucketName, id, edgeStack, func() {
		updateFunc(edgeStack)

		service.idxVersion[ID] = edgeStack.Version
		service.cacheInvalidationFn(ID)
	})
}

// UpdateEdgeStackFuncTx is a helper function used to call UpdateEdgeStackFunc inside a transaction.
func (service *Service) UpdateEdgeStackFuncTx(tx portainer.Transaction, ID portaineree.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error {
	return service.Tx(tx).UpdateEdgeStackFunc(ID, updateFunc)
}

// DeleteEdgeStack deletes an Edge stack.
func (service *Service) DeleteEdgeStack(ID portaineree.EdgeStackID) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.DeleteObject(BucketName, identifier)
	if err != nil {
		return err
	}

	delete(service.idxVersion, ID)

	service.cacheInvalidationFn(ID)

	return nil
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
