package services

import (
	"github.com/portainer/portainer-ee/api/docker/images"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/rs/zerolog/log"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

// ServiceImageStatus
// @id ServiceImageStatus
// @summary Fetch image status for service
// @description
// @description **Access policy**:
// @tags docker
// @security jwt
// @accept json
// @produce json
// @success 200 "Success"
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @router /docker/{environmentId}/services/{serviceID}/image_status [get]
func (handler *Handler) ServiceImageStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	serviceID, err := request.RetrieveRouteVariableValue(r, "serviceID")
	if err != nil {
		return httperror.BadRequest("Invalid serviceID", err)
	}

	s, err := images.CachedResourceImageStatus(serviceID)
	if err == nil {
		return response.JSON(w, &images.StatusResponse{Status: s, Message: ""})
	}

	log.Debug().Err(err).Str("serviceID", serviceID).Msg("No image status found from cache for service")

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.NotFound("Unable to find an environment on request context", err)
	}

	digestCli := images.NewClientWithRegistry(images.NewRegistryClient(handler.dataStore), handler.dockerClientFactory)

	s, err = digestCli.ServiceImageStatus(r.Context(), serviceID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to get the status of this image", err)
	}

	if s != images.Preparing {
		images.CacheResourceImageStatus(serviceID, s)
	}

	return response.JSON(w, &images.StatusResponse{Status: s, Message: ""})
}
