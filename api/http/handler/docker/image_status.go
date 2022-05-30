package docker

import (
	"github.com/portainer/portainer-ee/api/docker/images"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type imageStatusRequest struct {
	ImageName string `json:"ImageName"`
}

func (payload *imageStatusRequest) Validate(r *http.Request) error {
	return nil
}

func (handler *Handler) imageStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload imageStatusRequest
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.NotFound("Unable to find an environment on request context", err)
	}

	cli, err := handler.DockerClientFactory.CreateClient(endpoint, "", nil)
	if err != nil {
		return httperror.InternalServerError("Unable to connect to the Docker daemon", err)
	}

	digestCli := images.NewClientWithRegistry(images.NewRegistryClient(handler.DataStore), cli)

	s, err := digestCli.Status(r.Context(), payload.ImageName)
	if err != nil {
		return httperror.InternalServerError("Unable get the status of this image", err)
	}

	return response.JSON(w, &images.StatusResponse{Status: s, Message: ""})
}
