package edgestacks

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/set"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/pkg/featureflags"
	"github.com/rs/zerolog/log"
)

type updateEdgeStackPayload struct {
	StackFileContent string
	EdgeGroups       []portaineree.EdgeGroupID
	DeploymentType   portaineree.EdgeStackDeploymentType
	Registries       []portaineree.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
	PrePullImage          bool
	RePullImage           bool
	RetryDeploy           bool
	UpdateVersion         bool
	// Optional webhook configuration
	Webhook *string `example:"c11fdf23-183e-428a-9bb6-16db01032174"`
}

func (payload *updateEdgeStackPayload) Validate(r *http.Request) error {
	if payload.StackFileContent == "" {
		return errors.New("Invalid file content")
	}

	if len(payload.EdgeGroups) == 0 {
		return errors.New("edge groups are mandatory for an Edge stack")
	}

	return nil
}

// @id EdgeStackUpdate
// @summary Update an EdgeStack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "EdgeStack Id"
// @param body body updateEdgeStackPayload true "EdgeStack data"
// @success 200 {object} portaineree.EdgeStack
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id} [put]
func (handler *Handler) edgeStackUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	var payload updateEdgeStackPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var stack *portaineree.EdgeStack
	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		stack, err = handler.updateEdgeStack(handler.DataStore, portaineree.EdgeStackID(stackID), payload)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			stack, err = handler.updateEdgeStack(tx, portaineree.EdgeStackID(stackID), payload)
			return err
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, stack)
}

func (handler *Handler) updateEdgeStack(tx dataservices.DataStoreTx, stackID portaineree.EdgeStackID, payload updateEdgeStackPayload) (*portaineree.EdgeStack, error) {
	stack, err := tx.EdgeStack().EdgeStack(portaineree.EdgeStackID(stackID))
	if err != nil {
		return nil, handler.handlerDBErr(err, "Unable to find a stack with the specified identifier inside the database")
	}

	if stack.EdgeUpdateID != 0 {
		return nil, httperror.BadRequest("Unable to delete edge stack that is used by an edge update schedule", err)
	}

	relationConfig, err := edge.FetchEndpointRelationsConfig(tx)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve environments relations config from database", err)
	}

	relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(stack.EdgeGroups, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve edge stack related environments from database", err)
	}

	endpointsToAdd := set.Set[portaineree.EndpointID]{}
	groupsIds := stack.EdgeGroups
	if payload.EdgeGroups != nil {
		newRelated, newEndpoints, err := handler.handleChangeEdgeGroups(tx, stack.ID, payload.EdgeGroups, relatedEndpointIds, relationConfig)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to handle edge groups change", err)
		}

		groupsIds = payload.EdgeGroups
		relatedEndpointIds = newRelated
		endpointsToAdd = newEndpoints
	}

	entryPoint := stack.EntryPoint
	manifestPath := stack.ManifestPath
	deploymentType := stack.DeploymentType

	// shouldRemovePreviousVersion will be used when the deployment type is changed
	shouldRemovePreviousVersion := false
	if deploymentType != payload.DeploymentType {
		// deployment type was changed - need to delete all old files
		err = handler.FileService.RemoveDirectory(stack.ProjectPath)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to clear old files")
		}

		entryPoint = ""
		manifestPath = ""
		deploymentType = payload.DeploymentType

		shouldRemovePreviousVersion = true
	}

	hasWrongType, err := hasWrongEnvironmentType(tx.Endpoint(), relatedEndpointIds, payload.DeploymentType)
	if err != nil {
		return nil, httperror.BadRequest("unable to check for existence of non fitting environments: %w", err)
	}
	if hasWrongType {
		return nil, httperror.BadRequest("edge stack with config do not match the environment type", nil)
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	if deploymentType == portaineree.EdgeStackDeploymentCompose {
		if entryPoint == "" {
			entryPoint = filesystem.ComposeFileDefaultName
		}

		_, err = handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, entryPoint, stack.Version+1, []byte(payload.StackFileContent))
		if err != nil {
			return nil, httperror.InternalServerError("Unable to persist updated Compose file with version on disk", err)
		}
	}

	if deploymentType == portaineree.EdgeStackDeploymentKubernetes {
		if manifestPath == "" {
			manifestPath = filesystem.ManifestFileDefaultName
		}

		_, err = handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, manifestPath, stack.Version+1, []byte(payload.StackFileContent))
		if err != nil {
			return nil, httperror.InternalServerError("Unable to persist updated Kubernetes manifest with version on disk", err)
		}
	}

	if deploymentType == portaineree.EdgeStackDeploymentNomad {
		if entryPoint == "" {
			entryPoint = eefs.NomadJobFileDefaultName
		}

		_, err = handler.FileService.StoreEdgeStackFileFromBytesByVersion(stackFolder, entryPoint, stack.Version+1, []byte(payload.StackFileContent))
		if err != nil {
			return nil, httperror.InternalServerError("Unable to persist updated Nomad job file with version on disk", err)
		}
	}

	err = tx.EdgeStack().UpdateEdgeStackFunc(stack.ID, func(edgeStack *portaineree.EdgeStack) {
		// Assign a potentially new registries to the stack
		edgeStack.Registries = payload.Registries

		edgeStack.PrePullImage = payload.PrePullImage
		edgeStack.RePullImage = payload.RePullImage
		edgeStack.RetryDeploy = payload.RetryDeploy

		edgeStack.NumDeployments = len(relatedEndpointIds)
		if payload.UpdateVersion {
			if shouldRemovePreviousVersion {
				// if the deployment type is changed, we need to remove the previous version,
				// because the new version will be deployed with the new deployment type.
				// For example, if the deployment type was compose and now it's kubernetes,
				// there is no point to keep the old compose file version for kubernetes deployment type
				edgeStack.PreviousDeploymentInfo = nil
			} else {
				// When the edge stack is updated, we need to keep track of the previous version
				edgeStack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
					Version: edgeStack.Version,
				}
			}

			// Increment the stack version
			edgeStack.Version++

			edgeStack.Status = make(map[portaineree.EndpointID]portainer.EdgeStackStatus)
		}

		edgeStack.UseManifestNamespaces = payload.UseManifestNamespaces

		edgeStack.DeploymentType = deploymentType
		edgeStack.EntryPoint = entryPoint
		edgeStack.ManifestPath = manifestPath

		edgeStack.EdgeGroups = groupsIds

		if payload.Webhook != nil {
			edgeStack.Webhook = *payload.Webhook
		}
	})
	if err != nil {
		return nil, httperror.InternalServerError("Unable to persist the stack changes inside the database", err)
	}

	if payload.UpdateVersion {
		for _, endpointID := range relatedEndpointIds {
			endpoint, err := tx.Endpoint().Endpoint(endpointID)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to retrieve environment from the database", err)
			}

			if !endpointsToAdd[endpoint.ID] {
				err = handler.edgeAsyncService.ReplaceStackCommandTx(tx, endpoint, stack.ID)
				if err != nil {
					return nil, httperror.InternalServerError("Unable to store edge async command into the database", err)
				}
			}
		}
	}

	return stack, nil
}

func (handler *Handler) handleChangeEdgeGroups(tx dataservices.DataStoreTx, edgeStackID portaineree.EdgeStackID, newEdgeGroupsIDs []portaineree.EdgeGroupID, oldRelatedEnvironmentIDs []portaineree.EndpointID, relationConfig *edge.EndpointRelationsConfig) ([]portaineree.EndpointID, set.Set[portaineree.EndpointID], error) {
	newRelatedEnvironmentIDs, err := edge.EdgeStackRelatedEndpoints(newEdgeGroupsIDs, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Unable to retrieve edge stack related environments from database")
	}

	oldRelatedSet := set.ToSet(oldRelatedEnvironmentIDs)
	newRelatedSet := set.ToSet(newRelatedEnvironmentIDs)

	endpointsToRemove := set.Set[portaineree.EndpointID]{}
	for endpointID := range oldRelatedSet {
		if !newRelatedSet[endpointID] {
			endpointsToRemove[endpointID] = true
		}
	}

	for endpointID := range endpointsToRemove {
		relation, err := tx.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to find environment relation in database")
		}

		delete(relation.EdgeStacks, edgeStackID)

		err = tx.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to persist environment relation in database")
		}

		err = handler.edgeAsyncService.RemoveStackCommandTx(tx, endpointID, edgeStackID)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to store edge async command into the database")
		}
	}

	endpointsToAdd := set.Set[portaineree.EndpointID]{}
	for endpointID := range newRelatedSet {
		if !oldRelatedSet[endpointID] {
			endpointsToAdd[endpointID] = true
		}
	}

	for endpointID := range endpointsToAdd {
		relation, err := tx.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to find environment relation in database")
		}

		relation.EdgeStacks[edgeStackID] = true

		err = tx.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to persist environment relation in database")
		}

		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to retrieve environment from the database")
		}

		err = handler.edgeAsyncService.AddStackCommandTx(tx, endpoint, edgeStackID, "")
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to store edge async command into the database")
		}
	}

	return newRelatedEnvironmentIDs, endpointsToAdd, nil
}
