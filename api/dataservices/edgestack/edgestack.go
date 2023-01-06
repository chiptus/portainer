package edgestack

import (
	"fmt"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "edge_stack"
)

// Service represents a service for managing Edge stack data.
type Service struct {
	connection portainer.Connection
	idxVersion map[portaineree.EdgeStackID]int
	mu         sync.RWMutex
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
		idxVersion: make(map[portaineree.EdgeStackID]int),
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

// EdgeStacks returns an array containing all edge stacks
func (service *Service) EdgeStacks() ([]portaineree.EdgeStack, error) {
	var stacks = make([]portaineree.EdgeStack, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.EdgeStack{},
		func(obj interface{}) (interface{}, error) {
			stack, ok := obj.(*portaineree.EdgeStack)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to EdgeStack object")
				return nil, fmt.Errorf("Failed to convert to EdgeStack object: %s", obj)
			}

			stacks = append(stacks, *stack)

			return &portaineree.EdgeStack{}, nil
		})

	return stacks, err
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
	service.mu.Unlock()

	for endpointID := range edgeStack.Status {
		cache.Del(endpointID)
	}

	return nil
}

// Deprecated: Use UpdateEdgeStackFunc instead.
func (service *Service) UpdateEdgeStack(ID portaineree.EdgeStackID, edgeStack *portaineree.EdgeStack) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	prevEdgeStack, err := service.EdgeStack(ID)
	if err != nil {
		return err
	}

	identifier := service.connection.ConvertToKey(int(ID))

	err = service.connection.UpdateObject(BucketName, identifier, edgeStack)
	if err != nil {
		return err
	}

	service.idxVersion[ID] = edgeStack.Version

	// Invalidate cache for removed environments
	for endpointID := range prevEdgeStack.Status {
		if _, ok := edgeStack.Status[endpointID]; !ok {
			cache.Del(endpointID)
		}
	}

	// Invalidate cache when version changes and for added environments
	for endpointID := range edgeStack.Status {
		if prevEdgeStack.Version == edgeStack.Version {
			if _, ok := prevEdgeStack.Status[endpointID]; ok {
				continue
			}
		}

		cache.Del(endpointID)
	}

	return nil
}

// UpdateEdgeStackFunc updates an Edge stack inside a transaction avoiding data races.
func (service *Service) UpdateEdgeStackFunc(ID portaineree.EdgeStackID, updateFunc func(edgeStack *portaineree.EdgeStack)) error {
	id := service.connection.ConvertToKey(int(ID))
	edgeStack := &portaineree.EdgeStack{}

	service.mu.Lock()
	defer service.mu.Unlock()

	return service.connection.UpdateObjectFunc(BucketName, id, edgeStack, func() {
		prevEndpoints := make(map[portaineree.EndpointID]struct{}, len(edgeStack.Status))
		for endpointID := range edgeStack.Status {
			if _, ok := edgeStack.Status[endpointID]; !ok {
				prevEndpoints[endpointID] = struct{}{}
			}
		}

		updateFunc(edgeStack)

		prevVersion := service.idxVersion[ID]
		service.idxVersion[ID] = edgeStack.Version

		// Invalidate cache for removed environments
		for endpointID := range prevEndpoints {
			if _, ok := edgeStack.Status[endpointID]; !ok {
				cache.Del(endpointID)
			}
		}

		// Invalidate cache when version changes and for added environments
		for endpointID := range edgeStack.Status {
			if prevVersion == edgeStack.Version {
				if _, ok := prevEndpoints[endpointID]; ok {
					continue
				}
			}

			cache.Del(endpointID)
		}
	})
}

// DeleteEdgeStack deletes an Edge stack.
func (service *Service) DeleteEdgeStack(ID portaineree.EdgeStackID) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	edgeStack, err := service.EdgeStack(ID)
	if err != nil {
		return err
	}

	identifier := service.connection.ConvertToKey(int(ID))

	err = service.connection.DeleteObject(BucketName, identifier)
	if err != nil {
		return err
	}

	delete(service.idxVersion, ID)

	for endpointID := range edgeStack.Status {
		cache.Del(endpointID)
	}

	return nil
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
