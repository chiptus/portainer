package cloudcredentials

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
)

// @id cloudCredsGetAllDeprecated
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
// @deprecated
// @router /cloudcredentials [get]

// @id cloudCredsCreateDeprecated
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
// @deprecated
// @router /cloudcredentials [post]

// @id cloudCredsDeleteDeprecated
// @summary Delete a cloud credential
// @description Delete a cloud credential
// @description **Access policy**: authenticated
// @tags cloud_credentials
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id query string true "ID of the cloud credential"
// @success 200 {object} models.CloudCredential
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @deprecated
// @router /cloudcredentials/{id} [delete]

// @id cloudCredsUpdateDeprecated
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
// @param id path string true "ID of the cloud credential"
// @success 200 {object} models.CloudCredential
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @deprecated
// @router /cloudcredentials/{id} [put]

// @id cloudCredsGetByIDDeprecated
// @summary getByID gets a cloud credential by ID
// @description getByID gets a cloud credential by ID
// @description **Access policy**: authenticated
// @tags cloud_credentials
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path string true "ID of the cloud credential"
// @deprecated
// @success 200 {object} models.CloudCredential
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /cloud/credentials/{id} [get]

func deprecatedCloudCredentialsParser(w http.ResponseWriter, r *http.Request) (string, *httperror.HandlerError) {
	return "/cloud/credentials", nil
}

func deprecatedCloudCredentialsIdParser(w http.ResponseWriter, r *http.Request) (string, *httperror.HandlerError) {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return "", httperror.BadRequest("Invalid request. Unable to parse id route variable", err)
	}

	return fmt.Sprintf("/cloud/credentials/%d", id), nil
}
