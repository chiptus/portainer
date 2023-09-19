package cloudcredentials

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/database/models"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id cloudCredsDelete
// @summary delete delete a cloud credential by ID
// @description delete delete a cloud credential by ID
// @description **Access policy**: authenticated
// @tags cloud_credentials
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path string true "ID of the cloud credential"
// @success 200 {object} models.CloudCredential
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /cloud/credentials/{id} [post]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	id, _ := request.RetrieveNumericRouteVariableValue(r, "id")
	err := h.DataStore.CloudCredential().Delete(models.CloudCredentialID(id))
	if err != nil {
		return httperror.InternalServerError("Unable to delete cloud credential from the database", err)
	}

	return response.JSON(w, nil)
}
