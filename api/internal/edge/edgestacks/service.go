package edgestacks

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/edge/edgeasync"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Service represents a service for managing edge stacks.
type Service struct {
	dataStore        dataservices.DataStore
	edgeAsyncService *edgeasync.Service
}

// NewService returns a new instance of a service.
func NewService(dataStore dataservices.DataStore, edgeAsyncService *edgeasync.Service) *Service {

	return &Service{
		dataStore:        dataStore,
		edgeAsyncService: edgeAsyncService,
	}
}

type BuildEdgeStackArgs struct {
	Registries                []portaineree.RegistryID
	ScheduledTime             string
	UseManifestNamespaces     bool
	PrePullImage              bool
	RePullImage               bool
	RetryDeploy               bool
	SupportRelativePath       bool
	FilesystemPath            string
	SupportPerDeviceConfigs   bool
	PerDeviceConfigsMatchType portainer.PerDevConfigsFilterType
	PerDeviceConfigsPath      string
	EnvVars                   []portainer.Pair
}

// BuildEdgeStack builds the initial edge stack object
// PersistEdgeStack is required to be called after this to persist the edge stack
func (service *Service) BuildEdgeStack(
	tx dataservices.DataStoreTx,
	name string,
	deploymentType portaineree.EdgeStackDeploymentType,
	edgeGroups []portaineree.EdgeGroupID,
	args BuildEdgeStackArgs,
) (*portaineree.EdgeStack, error) {
	err := validateUniqueName(tx.EdgeStack().EdgeStacks, name)
	if err != nil {
		return nil, err
	}

	err = validateScheduledTime(args.ScheduledTime)
	if err != nil {
		return nil, err
	}

	stackID := tx.EdgeStack().GetNextIdentifier()
	return &portaineree.EdgeStack{
		ID:                        portaineree.EdgeStackID(stackID),
		Name:                      name,
		DeploymentType:            deploymentType,
		CreationDate:              time.Now().Unix(),
		EdgeGroups:                edgeGroups,
		Status:                    make(map[portaineree.EndpointID]portainer.EdgeStackStatus, 0),
		Version:                   1,
		Registries:                args.Registries,
		ScheduledTime:             args.ScheduledTime,
		UseManifestNamespaces:     args.UseManifestNamespaces,
		PrePullImage:              args.PrePullImage,
		RePullImage:               args.RePullImage,
		RetryDeploy:               args.RetryDeploy,
		SupportRelativePath:       args.SupportRelativePath,
		FilesystemPath:            args.FilesystemPath,
		SupportPerDeviceConfigs:   args.SupportPerDeviceConfigs,
		PerDeviceConfigsMatchType: args.PerDeviceConfigsMatchType,
		PerDeviceConfigsPath:      args.PerDeviceConfigsPath,
		EnvVars:                   args.EnvVars,
		StackFileVersion:          1,
	}, nil
}

func validateUniqueName(edgeStacksGetter func() ([]portaineree.EdgeStack, error), name string) error {
	edgeStacks, err := edgeStacksGetter()
	if err != nil {
		return err
	}

	for _, stack := range edgeStacks {
		if strings.EqualFold(stack.Name, name) {
			return httperrors.NewConflictError("Edge stack name must be unique")
		}
	}

	return nil
}

func validateScheduledTime(scheduledTime string) error {
	// scheduled time is not required
	if scheduledTime == "" {
		return nil
	}

	parsedScheduledTime, err := time.Parse(portaineree.DateTimeFormat, string(scheduledTime))
	if err != nil {
		return errors.WithMessage(err, "invalid scheduled time")
	}

	if parsedScheduledTime.Before(time.Now().Add(-24 * time.Hour)) {
		return errors.New("scheduled time must be at most 24 hours in the past")
	}

	return nil
}

// PersistEdgeStack persists the edge stack in the database and its relations
func (service *Service) PersistEdgeStack(
	tx dataservices.DataStoreTx,
	stack *portaineree.EdgeStack,
	storeManifest edgetypes.StoreManifestFunc,
) (*portaineree.EdgeStack, error) {
	relationConfig, err := edge.FetchEndpointRelationsConfig(service.dataStore)
	if err != nil {
		return nil, fmt.Errorf("unable to find environment relations in database: %w", err)
	}

	relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(stack.EdgeGroups, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		if errors.Is(err, edge.ErrEdgeGroupNotFound) {
			return nil, httperrors.NewInvalidPayloadError(err.Error())
		}
		return nil, fmt.Errorf("unable to persist environment relation in database: %w", err)
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	configPath, manifestPath, projectPath, err := storeManifest(stackFolder, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("unable to store manifest: %w", err)
	}

	stack.ManifestPath = manifestPath
	stack.ProjectPath = projectPath
	stack.EntryPoint = configPath
	stack.NumDeployments = len(relatedEndpointIds)
	stack.Status = NewStatus(nil, relatedEndpointIds)

	err = service.updateEndpointRelations(tx, stack.ID, relatedEndpointIds)
	if err != nil {
		return nil, fmt.Errorf("unable to update endpoint relations: %w", err)
	}

	err = tx.EdgeStack().Create(stack.ID, stack)
	if err != nil {
		return nil, err
	}

	err = service.createEdgeCommands(tx, stack.ID, relatedEndpointIds, stack.ScheduledTime)
	if err != nil {
		return nil, fmt.Errorf("unable to update environment relations: %w", err)
	}

	return stack, nil
}

// updateEndpointRelations adds a relation between the Edge Stack to the related environments(endpoints)
func (service *Service) updateEndpointRelations(tx dataservices.DataStoreTx, edgeStackID portaineree.EdgeStackID, relatedEndpointIds []portaineree.EndpointID) error {
	for _, endpointID := range relatedEndpointIds {
		relation, err := tx.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return fmt.Errorf("unable to find endpoint relation in database: %w", err)
		}

		relation.EdgeStacks[edgeStackID] = true

		err = tx.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return fmt.Errorf("unable to persist endpoint relation in database: %w", err)
		}
	}

	return nil
}

func (service *Service) createEdgeCommands(tx dataservices.DataStoreTx, edgeStackID portaineree.EdgeStackID, relatedEndpointIds []portaineree.EndpointID, scheduledTime string) error {
	for _, endpointID := range relatedEndpointIds {
		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return err
		}

		err = service.edgeAsyncService.AddStackCommandTx(tx, endpoint, edgeStackID, scheduledTime)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteEdgeStack deletes the edge stack from the database and its relations
func (service *Service) DeleteEdgeStack(tx dataservices.DataStoreTx, edgeStackID portaineree.EdgeStackID, relatedEdgeGroupsIds []portaineree.EdgeGroupID) error {
	relationConfig, err := edge.FetchEndpointRelationsConfig(tx)
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve environments relations config from database")
	}

	relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(relatedEdgeGroupsIds, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		return errors.WithMessage(err, "Unable to retrieve edge stack related environments from database")
	}

	for _, endpointID := range relatedEndpointIds {
		relation, err := tx.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			if tx.IsErrObjectNotFound(err) {
				log.Warn().
					Int("endpoint_id", int(endpointID)).
					Msg("Unable to find endpoint relation in database, skipping")
				continue
			}

			return errors.WithMessage(err, "Unable to find environment relation in database")
		}

		delete(relation.EdgeStacks, edgeStackID)

		err = tx.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return errors.WithMessage(err, "Unable to persist environment relation in database")
		}

		err = service.edgeAsyncService.RemoveStackCommandTx(tx, endpointID, edgeStackID)
		if err != nil {
			return errors.WithMessage(err, "Unable to store edge async command into the database")
		}
	}

	err = tx.EdgeStack().DeleteEdgeStack(portaineree.EdgeStackID(edgeStackID))
	if err != nil {
		return errors.WithMessage(err, "Unable to remove the edge stack from the database")
	}

	return nil
}
