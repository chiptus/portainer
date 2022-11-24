package registries

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
)

type registryCreatePayload struct {
	// Name that will be used to identify this registry
	Name string `example:"my-registry" validate:"required"`
	// Registry Type. Valid values are:
	//	1 (Quay.io),
	//	2 (Azure container registry),
	//	3 (custom registry),
	//	4 (Gitlab registry),
	//	5 (ProGet registry),
	//	6 (DockerHub)
	//	7 (ECR)
	//	8 (Github registry)
	Type portaineree.RegistryType `example:"1" validate:"required" enums:"1,2,3,4,5,6,7,8"`
	// URL or IP address of the Docker registry
	URL string `example:"registry.mydomain.tld:2375/feed" validate:"required"`
	// BaseURL required for ProGet registry
	BaseURL string `example:"registry.mydomain.tld:2375"`
	// Is authentication against this registry enabled
	Authentication bool `example:"false" validate:"required"`
	// Username used to authenticate against this registry. Required when Authentication is true
	Username string `example:"registry_user"`
	// Password used to authenticate against this registry. required when Authentication is true
	Password string `example:"registry_password"`
	// Gitlab specific details, required when type = 4
	Gitlab portaineree.GitlabRegistryData
	// Quay specific details, required when type = 1
	Quay portaineree.QuayRegistryData
	// Github specific details, required when type = 8
	Github portaineree.GithubRegistryData
	// ECR specific details, required when type = 7
	Ecr portaineree.EcrData
}

func (payload *registryCreatePayload) Validate(_ *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid registry name")
	}
	if govalidator.IsNull(payload.URL) {
		return errors.New("Invalid registry URL")
	}

	if payload.Authentication {
		if govalidator.IsNull(payload.Username) || govalidator.IsNull(payload.Password) {
			return errors.New("Invalid credentials. Username and password must be specified when authentication is enabled")
		}
		if payload.Type == portaineree.EcrRegistry {
			if govalidator.IsNull(payload.Ecr.Region) {
				return errors.New("invalid credentials: access key ID, secret access key and region must be specified when authentication is enabled")
			}
		}
	}

	switch payload.Type {
	case portaineree.QuayRegistry, portaineree.AzureRegistry, portaineree.CustomRegistry, portaineree.GitlabRegistry, portaineree.ProGetRegistry, portaineree.DockerHubRegistry, portaineree.EcrRegistry, portaineree.GithubRegistry:
	default:
		return errors.New("invalid registry type. Valid values are: 1 (Quay.io), 2 (Azure container registry), 3 (custom registry), 4 (Gitlab registry), 5 (ProGet registry), 6 (DockerHub), 7 (ECR), 8 (Github registry)")
	}

	if payload.Type == portaineree.ProGetRegistry && payload.BaseURL == "" {
		return fmt.Errorf("BaseURL is required for registry type %d (ProGet)", portaineree.ProGetRegistry)
	}

	return nil
}

// @id RegistryCreate
// @summary Create a new registry
// @description Create a new registry.
// @description **Access policy**: restricted
// @tags registries
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body registryCreatePayload true "Registry details"
// @success 200 {object} portaineree.Registry "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /registries [post]
func (handler *Handler) registryCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}
	if !securityContext.IsAdmin {
		return httperror.Forbidden("Permission denied to create registry", httperrors.ErrResourceAccessDenied)
	}

	var payload registryCreatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	registry := &portaineree.Registry{
		Type:             portaineree.RegistryType(payload.Type),
		Name:             payload.Name,
		URL:              payload.URL,
		BaseURL:          payload.BaseURL,
		Authentication:   payload.Authentication,
		Username:         payload.Username,
		Password:         payload.Password,
		Gitlab:           payload.Gitlab,
		Quay:             payload.Quay,
		Github:           payload.Github,
		RegistryAccesses: portaineree.RegistryAccesses{},
		Ecr:              payload.Ecr,
	}

	registries, err := handler.DataStore.Registry().Registries()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve registries from the database", err)
	}
	for _, r := range registries {
		if r.Name == registry.Name {
			return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: "Another registry with the same name already exists", Err: errors.New("A registry is already defined with this name")}
		}
		if handler.registriesHaveSameURLAndCredentials(&r, registry) {
			return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: "Another registry with the same URL and credentials already exists", Err: errors.New("A registry is already defined for this URL and credentials")}
		}
	}

	err = handler.DataStore.Registry().Create(registry)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the registry inside the database", err)
	}

	hideFields(registry, true)
	return response.JSON(w, registry)
}
