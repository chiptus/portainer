package endpoints

import (
	"context"
	"net/http"
	"strings"

	dockertypes "github.com/docker/docker/api/types"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

type forceUpdateServicePayload struct {
	// ServiceId to update
	ServiceID string
	// PullImage if true will pull the image
	PullImage bool
}

func (payload *forceUpdateServicePayload) Validate(r *http.Request) error {
	return nil
}

// @id endpointForceUpdateService
// @summary force update a docker service
// @description force update a docker service
// @description **Access policy**: authenticated
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "endpoint identifier"
// @param body body forceUpdateServicePayload true "details"
// @success 200 {object} dockertypes.ServiceUpdateResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "endpoint not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/forceupdateservice [put]
func (handler *Handler) endpointForceUpdateService(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment identifier route variable", err}
	}

	var payload forceUpdateServicePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to force update service", err}
	}

	dockerClient, err := handler.DockerClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error creating docker client", err}
	}
	defer dockerClient.Close()

	service, _, err := dockerClient.ServiceInspectWithRaw(context.Background(), payload.ServiceID, dockertypes.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error looking up service", err}
	}

	service.Spec.TaskTemplate.ForceUpdate++

	if payload.PullImage {
		service.Spec.TaskTemplate.ContainerSpec.Image = strings.Split(service.Spec.TaskTemplate.ContainerSpec.Image, "@sha")[0]
	}

	newService, err := dockerClient.ServiceUpdate(context.Background(), payload.ServiceID, service.Version, service.Spec, dockertypes.ServiceUpdateOptions{QueryRegistry: true})
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error force update service", err}
	}

	return response.JSON(w, newService)
}
