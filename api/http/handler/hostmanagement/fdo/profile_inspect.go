package fdo

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type fdoProfileResponse struct {
	Name        string `json:"name"`
	FileContent string `json:"fileContent"`
}

// @id fdoProfileInspect
// @summary retrieves a given FDO profile information and content
// @description retrieves a given FDO profile information and content
// @description **Access policy**: administrator
// @tags intel
// @security jwt
// @produce json
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /fdo/profiles/{id} [get]
func (handler *Handler) fdoProfileInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Bad request", errors.New("missing 'id' query parameter"))
	}

	profile, err := handler.DataStore.FDOProfile().FDOProfile(portaineree.FDOProfileID(id))
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Profile", err)
	}

	fileContent, err := handler.FileService.GetFileContent(profile.FilePath, "")
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Profile file content", err)
	}

	return response.JSON(w, fdoProfileResponse{
		Name:        profile.Name,
		FileContent: string(fileContent),
	})
}
