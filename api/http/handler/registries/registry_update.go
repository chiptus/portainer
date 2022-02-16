package registries

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

type registryUpdatePayload struct {
	Name             *string `json:",omitempty" example:"my-registry" validate:"required"`
	URL              *string `json:",omitempty" example:"registry.mydomain.tld:2375/feed" validate:"required"`
	BaseURL          *string `json:",omitempty" example:"registry.mydomain.tld:2375"`
	Authentication   *bool   `json:",omitempty" example:"false" validate:"required"`
	Username         *string `json:",omitempty" example:"registry_user"`
	Password         *string `json:",omitempty" example:"registry_password"`
	Quay             *portaineree.QuayRegistryData
	RegistryAccesses *portaineree.RegistryAccesses `json:",omitempty"`
	Ecr              *portaineree.EcrData          `json:",omitempty" example:"\{\"Region\": \"ap-southeast-2\"\}"`
}

func (payload *registryUpdatePayload) Validate(r *http.Request) error {
	return nil
}

// @id RegistryUpdate
// @summary Update a registry
// @description Update a registry
// @description **Access policy**: restricted
// @tags registries
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Registry identifier"
// @param body body registryUpdatePayload true "Registry details"
// @success 200 {object} portaineree.Registry "Success"
// @failure 400 "Invalid request"
// @failure 404 "Registry not found"
// @failure 409 "Another registry with the same URL already exists"
// @failure 500 "Server error"
// @router /registries/{id} [put]
func (handler *Handler) registryUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve info from request context", err}
	}

	if !securityContext.IsAdmin {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to update registry", httperrors.ErrResourceAccessDenied}
	}

	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid registry identifier route variable", err}
	}

	registry, err := handler.DataStore.Registry().Registry(portaineree.RegistryID(registryID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a registry with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a registry with the specified identifier inside the database", err}
	}

	var payload registryUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	if payload.Name != nil {
		registry.Name = *payload.Name
	}

	if registry.Type == portaineree.ProGetRegistry && payload.BaseURL != nil {
		registry.BaseURL = *payload.BaseURL
	}

	shouldUpdateSecrets := false

	if payload.Authentication != nil {
		shouldUpdateSecrets = shouldUpdateSecrets || (registry.Authentication != *payload.Authentication)

		if *payload.Authentication {
			registry.Authentication = true

			if payload.Username != nil {
				shouldUpdateSecrets = shouldUpdateSecrets || (registry.Username != *payload.Username)
				registry.Username = *payload.Username
			}

			if payload.Password != nil && *payload.Password != "" {
				shouldUpdateSecrets = shouldUpdateSecrets || (registry.Password != *payload.Password)
				registry.Password = *payload.Password
			}

			if registry.Type == portaineree.EcrRegistry && payload.Ecr != nil && payload.Ecr.Region != "" {
				shouldUpdateSecrets = shouldUpdateSecrets || (registry.Ecr.Region != payload.Ecr.Region)
				registry.Ecr.Region = payload.Ecr.Region
			}
		} else {
			registry.Authentication = false
			registry.Username = ""
			registry.Password = ""

			registry.Ecr.Region = ""

			registry.AccessToken = ""
			registry.AccessTokenExpiry = 0
		}
	}

	if payload.URL != nil {
		shouldUpdateSecrets = shouldUpdateSecrets || (*payload.URL != registry.URL)

		registry.URL = *payload.URL
		registries, err := handler.DataStore.Registry().Registries()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve registries from the database", err}
		}

		for _, r := range registries {
			if r.ID != registry.ID && handler.registriesHaveSameURLAndCredentials(&r, registry) {
				return &httperror.HandlerError{http.StatusConflict, "Another registry with the same URL and credentials already exists", errors.New("A registry is already defined for this URL and credentials")}
			}
		}
	}

	if shouldUpdateSecrets {
		registry.AccessToken = ""
		registry.AccessTokenExpiry = 0

		for endpointID, endpointAccess := range registry.RegistryAccesses {
			endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update access to registry", err}
			}

			if endpointutils.IsKubernetesEndpoint(endpoint) {
				err = handler.updateEndpointRegistryAccess(endpoint, registry, endpointAccess)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update access to registry", err}
				}
			}
		}
	}

	if payload.Quay != nil {
		registry.Quay = *payload.Quay
	}

	err = handler.DataStore.Registry().UpdateRegistry(registry.ID, registry)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist registry changes inside the database", err}
	}

	return response.JSON(w, registry)
}

func (handler *Handler) updateEndpointRegistryAccess(endpoint *portaineree.Endpoint, registry *portaineree.Registry, endpointAccess portaineree.RegistryAccessPolicies) error {

	cli, err := handler.K8sClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return err
	}

	for _, namespace := range endpointAccess.Namespaces {
		err := cli.DeleteRegistrySecret(registry, namespace)
		if err != nil {
			return err
		}

		err = cli.CreateRegistrySecret(registry, namespace)
		if err != nil {
			return err
		}
	}

	return nil
}
