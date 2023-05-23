package edgestacks

import (
	"errors"
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/pkg/featureflags"

	"github.com/rs/zerolog/log"
)

type updateEdgeStackPayload struct {
	StackFileContent string
	Version          *int
	EdgeGroups       []portaineree.EdgeGroupID
	DeploymentType   portaineree.EdgeStackDeploymentType
	Registries       []portaineree.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
	PrePullImage          bool
	RePullImage           bool
	RetryDeploy           bool
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

	endpointsToAdd := map[portaineree.EndpointID]bool{}

	if payload.EdgeGroups != nil {
		newRelated, err := edge.EdgeStackRelatedEndpoints(payload.EdgeGroups, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve edge stack related environments from database", err)
		}

		oldRelatedSet := endpointutils.EndpointSet(relatedEndpointIds)
		newRelatedSet := endpointutils.EndpointSet(newRelated)

		endpointsToRemove := map[portaineree.EndpointID]bool{}
		for endpointID := range oldRelatedSet {
			if !newRelatedSet[endpointID] {
				endpointsToRemove[endpointID] = true
			}
		}

		for endpointID := range endpointsToRemove {
			relation, err := tx.EndpointRelation().EndpointRelation(endpointID)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to find environment relation in database", err)
			}

			delete(relation.EdgeStacks, stack.ID)

			err = tx.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to persist environment relation in database", err)
			}

			err = handler.edgeAsyncService.RemoveStackCommandTx(tx, endpointID, stack.ID)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to store edge async command into the database", err)
			}
		}

		for endpointID := range newRelatedSet {
			if !oldRelatedSet[endpointID] {
				endpointsToAdd[endpointID] = true
			}
		}

		for endpointID := range endpointsToAdd {
			relation, err := tx.EndpointRelation().EndpointRelation(endpointID)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to find environment relation in database", err)
			}

			relation.EdgeStacks[stack.ID] = true

			err = tx.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to persist environment relation in database", err)
			}

			endpoint, err := tx.Endpoint().Endpoint(endpointID)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to retrieve environment from the database", err)
			}

			err = handler.edgeAsyncService.AddStackCommandTx(tx, endpoint, stack.ID, "")
			if err != nil {
				return nil, httperror.InternalServerError("Unable to store edge async command into the database", err)
			}
		}

		stack.EdgeGroups = payload.EdgeGroups
		relatedEndpointIds = newRelated
	}

	if stack.DeploymentType != payload.DeploymentType {
		// deployment type was changed - need to delete the old file
		err = handler.FileService.RemoveDirectory(stack.ProjectPath)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to clear old files")
		}

		stack.EntryPoint = ""
		stack.ManifestPath = ""
		stack.DeploymentType = payload.DeploymentType
	}

	stackFolder := strconv.Itoa(int(stack.ID))

	hasWrongType, err := hasWrongEnvironmentType(tx.Endpoint(), relatedEndpointIds, payload.DeploymentType)
	if err != nil {
		return nil, httperror.BadRequest("unable to check for existence of non fitting environments: %w", err)
	}
	if hasWrongType {
		return nil, httperror.BadRequest("edge stack with config do not match the environment type", nil)
	}

	if payload.DeploymentType == portaineree.EdgeStackDeploymentCompose {
		if stack.EntryPoint == "" {
			stack.EntryPoint = filesystem.ComposeFileDefaultName
		}

		_, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
		if err != nil {
			return nil, httperror.InternalServerError("Unable to persist updated Compose file on disk", err)
		}

		manifestPath, err := handler.convertAndStoreKubeManifestIfNeeded(stackFolder, stack.ProjectPath, stack.EntryPoint, relatedEndpointIds)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to convert and persist updated Kubernetes manifest file on disk", err)
		}

		stack.ManifestPath = manifestPath
	}

	if payload.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
		if stack.ManifestPath == "" {
			stack.ManifestPath = filesystem.ManifestFileDefaultName
		}

		stack.UseManifestNamespaces = payload.UseManifestNamespaces

		_, err = handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.ManifestPath, []byte(payload.StackFileContent))
		if err != nil {
			return nil, httperror.InternalServerError("Unable to persist updated Kubernetes manifest file on disk", err)
		}
	}

	if payload.DeploymentType == portaineree.EdgeStackDeploymentNomad {
		if stack.EntryPoint == "" {
			stack.EntryPoint = eefs.NomadJobFileDefaultName
		}

		_, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
		if err != nil {
			return nil, httperror.InternalServerError("Unable to persist updated Nomad job file on disk", err)
		}
	}

	versionUpdated := payload.Version != nil && *payload.Version != stack.Version
	if versionUpdated {
		stack.Version = *payload.Version
	}

	// Assign a potentially new registries to the stack
	stack.Registries = payload.Registries

	stack.PrePullImage = payload.PrePullImage
	stack.RePullImage = payload.RePullImage
	stack.RetryDeploy = payload.RetryDeploy

	stack.NumDeployments = len(relatedEndpointIds)

	if versionUpdated {
		stack.Status = make(map[portaineree.EndpointID]portainer.EdgeStackStatus)
	}

	err = tx.EdgeStack().UpdateEdgeStack(stack.ID, stack)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to persist the stack changes inside the database", err)
	}

	if !versionUpdated {
		return stack, nil
	}

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

	return stack, nil
}
