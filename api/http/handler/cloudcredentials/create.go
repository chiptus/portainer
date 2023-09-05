package cloudcredentials

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/database/models"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id Create
// @summary Create a cloud credential
// @description Create a cloud credential
// @description **Access policy**: authenticated
// @tags cloud_credentials
// @security ApiKeyAuth
// @security jwt
// @accept json,multipart/form-data
// @produce json
// @param provider formData string true "cloud provider such as aws, aks, civo, digitalocean, etc."
// @param name formData string true "name of the credentials such as rnd-test-credential"
// @param credentials formData string true "credentials in json format"
// @success 200 {object} models.CloudCredential
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /cloud/credentials [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload models.CloudCredential
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	cloudCredentials, err := h.DataStore.CloudCredential().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to validate credential name", err)
	}
	err = payload.ValidateUniqueNameByProvider(cloudCredentials)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	err = h.DataStore.CloudCredential().Create(&payload)
	if err != nil {
		return httperror.InternalServerError("Unable to persist settings changes inside the database", err)
	}

	payload.Credentials = redactCredentials(payload.Credentials)

	return response.JSON(w, payload)
}
