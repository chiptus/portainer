package cloudcredentials

import (
	"net/http"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/database/models"
)

var redactedKeys = []string{"jsonKeyBase64", "apiKey", "secretAccessKey", "clientSecret"}

func redactCredentials(credential models.CloudCredentialMap) models.CloudCredentialMap {
	if credential == nil {
		return nil
	}

	for _, key := range redactedKeys {
		if val, ok := credential[key]; ok {
			credential[key] = strings.Repeat("*", len(val))
		}
	}

	return credential
}

// @id getAll
// @summary getAll cloud credentials
// @description getAll cloud credential
// @description **Access policy**: authenticated
// @tags cloud_credentials
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {object} models.CloudCredential
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /cloudcredentials [get]
func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	cloudCredentials, err := h.DataStore.CloudCredential().GetAll()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to fetch cloud credentials from the database", Err: err}
	}

	for i, cred := range cloudCredentials {
		cloudCredentials[i].Credentials = redactCredentials(cred.Credentials)
	}

	return response.JSON(w, cloudCredentials)
}

// @id getByID
// @summary getByID gets a cloud credential by ID
// @description getByID gets a cloud credential by ID
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
func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	id, _ := request.RetrieveNumericRouteVariableValue(r, "id")
	cloudCredential, err := h.DataStore.CloudCredential().GetByID(models.CloudCredentialID(id))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to fetch cloud credentials from the database", Err: err}
	}

	cloudCredential.Credentials = redactCredentials(cloudCredential.Credentials)

	return response.JSON(w, cloudCredential)
}
