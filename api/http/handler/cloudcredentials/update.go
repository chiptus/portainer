package cloudcredentials

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
)

// @id Update
// @summary Update a cloud credential
// @description Update a cloud credential
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
// @router /cloudcredentials [put]
func (h *Handler) update(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	id, _ := request.RetrieveNumericRouteVariableValue(r, "id")
	cloudCredential, err := h.DataStore.CloudCredential().GetByID(models.CloudCredentialID(id))
	if err != nil {
		return httperror.InternalServerError("Unable to fetch cloud credentials from the database", err)
	}

	if cloudCredential.Provider == portaineree.CloudProviderKubeConfig {
		return httperror.BadRequest("Invalid request", err)
	}

	var payload models.CloudCredential
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	if payload.Name != "" {
		cloudCredential.Name = payload.Name
	}

	if payload.Provider != "" {
		cloudCredential.Provider = payload.Provider
	}

	if payload.Credentials != nil {
		credentials := cloudCredential.Credentials
		for k, v := range payload.Credentials {
			credentials[k] = v
		}
		cloudCredential.Credentials = credentials
	}

	cloudCredentials, err := h.DataStore.CloudCredential().GetAll()
	if err != nil {
		return httperror.InternalServerError("Unable to validate credential name", err)
	}
	err = cloudCredential.ValidateUniqueNameByProvider(cloudCredentials)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	err = h.DataStore.CloudCredential().Update(models.CloudCredentialID(id), cloudCredential)
	if err != nil {
		return httperror.InternalServerError("Unable to update cloud credential in the database", err)
	}

	cloudCredential.Credentials = redactCredentials(cloudCredential.Credentials)
	return response.JSON(w, cloudCredential)
}
