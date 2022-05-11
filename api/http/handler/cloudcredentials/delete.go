package cloudcredentials

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/database/models"
)

// @id delete
// @summary delete delete a cloud credential by ID
// @description delete delete a cloud credential by ID
// @description **Access policy**: authenticated
// @tags cloud_credentials
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id query string true "ID of the cloud credential"
// @success 200 {object} models.CloudCredential
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /cloudcredentials [get]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	id, _ := request.RetrieveNumericRouteVariableValue(r, "id")
	err := h.DataStore.CloudCredential().Delete(models.CloudCredentialID(id))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to delete cloud credential from the database", Err: err}
	}

	return response.JSON(w, nil)
}
