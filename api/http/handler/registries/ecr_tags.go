package registries

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	ecr "github.com/portainer/portainer-ee/api/aws/ecr"
	"github.com/portainer/portainer-ee/api/internal/registryutils"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

type deleteTagsPayload struct {
	Tags []string
}

func (p *deleteTagsPayload) Validate(r *http.Request) error {
	return nil
}

// @id ecrDeleteTags
// @summary Delete tags
// @description Delete tags for a given ECR repository
// @description **Access policy**: restricted
// @tags registries
// @security jwt
// @param id path int true "Registry identifier"
// @param body body deleteTagsPayload true "Tag Array"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access registry"
// @failure 404 "Registry not found"
// @failure 500 "Server error"
// @router /registries/{id}/ecr/repositories/{repositoryName}/tags [delete]
func (handler *Handler) ecrDeleteTags(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid registry identifier route variable", err}
	}

	repositoryName, err := request.RetrieveRouteVariableValue(r, "repositoryName")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid repository name route variable", err}
	}

	var payload deleteTagsPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	registry, err := handler.DataStore.Registry().Registry(portaineree.RegistryID(registryID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a registry with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a registry with the specified identifier inside the database", err}
	}

	username, password, region := registryutils.GetManagementCredential(registry)
	ecrClient := ecr.NewService(username, password, region)

	registryId, err := registryutils.GetRegistryId(registry)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to get registry ID", err}
	}

	err = ecrClient.DeleteTags(&registryId, &repositoryName, payload.Tags)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to delete ECR tags", err}
	}

	return response.Empty(w)
}
