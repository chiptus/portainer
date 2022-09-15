package endpointedge

import (
	"errors"
	"net/http"

	"github.com/docker/docker/api/types"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
)

type EdgeAsyncContainerCommandCreateRequest struct {
	ContainerName          string
	ContainerStartOptions  types.ContainerStartOptions
	ContainerRemoveOptions types.ContainerRemoveOptions
	ContainerOperation     portaineree.EdgeAsyncContainerOperation
}

type EdgeAsyncImageCommandCreateRequest struct {
	ImageName          string
	ImageOperation     portaineree.EdgeAsyncImageOperation
	ImageRemoveOptions types.ImageRemoveOptions
}

type EdgeAsyncVolumeCommandCreateRequest struct {
	VolumeName      string
	VolumeOperation portaineree.EdgeAsyncVolumeOperation
	ForceRemove     bool
}

func (payload *EdgeAsyncContainerCommandCreateRequest) Validate(r *http.Request) error {
	if len(payload.ContainerName) == 0 {
		return errors.New("container name is mandatory")
	}

	if len(payload.ContainerOperation) == 0 {
		return errors.New("container operation is mandatory")
	}

	return nil
}

func (payload *EdgeAsyncImageCommandCreateRequest) Validate(r *http.Request) error {
	if len(payload.ImageName) == 0 {
		return errors.New("image name is mandatory")
	}

	if len(payload.ImageOperation) == 0 {
		return errors.New("image operation is mandatory")
	}

	return nil
}

func (payload *EdgeAsyncVolumeCommandCreateRequest) Validate(r *http.Request) error {
	if len(payload.VolumeName) == 0 {
		return errors.New("volume name is mandatory")
	}

	if len(payload.VolumeOperation) == 0 {
		return errors.New("volume operation is mandatory")
	}

	return nil
}

func (handler *Handler) createContainerCommand(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	var payload EdgeAsyncContainerCommandCreateRequest
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	switch payload.ContainerOperation {
	case portaineree.EdgeAsyncContainerOperationStart:
		err = handler.EdgeService.StartContainerCommand(endpoint.ID, payload.ContainerName, payload.ContainerStartOptions)
	case portaineree.EdgeAsyncContainerOperationRestart:
		err = handler.EdgeService.RestartContainerCommand(endpoint.ID, payload.ContainerName)
	case portaineree.EdgeAsyncContainerOperationStop:
		err = handler.EdgeService.StopContainerCommand(endpoint.ID, payload.ContainerName)
	case portaineree.EdgeAsyncContainerOperationDelete:
		err = handler.EdgeService.DeleteContainerCommand(endpoint.ID, payload.ContainerName, payload.ContainerRemoveOptions)
	case portaineree.EdgeAsyncContainerOperationKill:
		err = handler.EdgeService.KillContainerCommand(endpoint.ID, payload.ContainerName)
	}

	if err != nil {
		return httperror.InternalServerError("Unable to create edge async container command", nil)
	}

	return response.JSON(w, []string{})
}

func (handler *Handler) createImageCommand(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	var payload EdgeAsyncImageCommandCreateRequest
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", nil)
	}

	switch payload.ImageOperation {
	case portaineree.EdgeAsyncImageOperationDelete:
		handler.EdgeService.DeleteImageCommand(endpoint.ID, payload.ImageName, payload.ImageRemoveOptions)
	}

	if err != nil {
		return httperror.InternalServerError("Unable to create edge async image command", nil)
	}

	return response.JSON(w, []string{})
}

func (handler *Handler) createVolumeCommand(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	var payload EdgeAsyncVolumeCommandCreateRequest
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", nil)
	}

	switch payload.VolumeOperation {
	case portaineree.EdgeAsyncVolumeOperationDelete:
		err = handler.EdgeService.DeleteVolumeCommand(endpoint.ID, payload.VolumeName, payload.ForceRemove)
	}

	if err != nil {
		return httperror.InternalServerError("Unable to create edge async volume command", nil)
	}

	return response.JSON(w, []string{})
}
