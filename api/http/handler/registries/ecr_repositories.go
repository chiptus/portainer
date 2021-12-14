package registries

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	ecr "github.com/portainer/portainer/api/aws/ecr"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/internal/registryutils"
)

// @id ecrDeleteRepository
// @summary Delete ECR repository
// @description Delete ECR repository.
// @description **Access policy**: restricted
// @tags registries
// @security jwt
// @produce json
// @param id path int true "Registry identifier"
// @param repositoryName string true "Repository name"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access registry"
// @failure 404 "Registry not found"
// @failure 500 "Server error"
// @router /registries/{id}/ecr/repositories/{repositoryName} [delete]
func (handler *Handler) ecrDeleteRepository(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid registry identifier route variable", err}
	}

	repositoryName, err := request.RetrieveRouteVariableValue(r, "repositoryName")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid repository name route variable", err}
	}

	registry, err := handler.DataStore.Registry().Registry(portainer.RegistryID(registryID))
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

	err = ecrClient.DeleteRepository(&registryId, &repositoryName)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to list ECR repositories", err}
	}

	return response.Empty(w)
}
