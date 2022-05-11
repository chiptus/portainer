package cloudcredentials

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/database/models"
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
// @router /cloudcredentials [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload models.CloudCredential
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	cloudCredentials, err := h.DataStore.CloudCredential().GetAll()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to validate credential name", Err: err}
	}
	err = payload.ValidateUniqueNameByProvider(cloudCredentials)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	err = h.DataStore.CloudCredential().Create(&payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist settings changes inside the database", Err: err}
	}

	payload.Credentials = redactCredentials(payload.Credentials)

	return response.JSON(w, payload)
}
