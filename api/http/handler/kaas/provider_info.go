package kaas

import (
	"errors"
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
)

// @id kaasProviderInfo
// @summary Returns information about a Cloud provider.
// @description The information returned can be used to provision a KaaS cluster.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud [post]
func (handler *Handler) kaasProviderInfo(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	provider, err := request.RetrieveRouteVariableValue(r, "provider")
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid user identifier route variable",
			Err:        err,
		}
	}

	force, err := request.RetrieveBooleanQueryParameter(r, "force", true)
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Failed parsing \"force\" boolean",
			Err:        err,
		}
	}

	credentialId, _ := request.RetrieveNumericQueryParameter(r, "credentialId", true)
	if credentialId == 0 {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Missing credential id in the query parameter",
			Err:        err,
		}
	}

	credential, err := handler.DataStore.CloudCredential().GetByID(models.CloudCredentialID(credentialId))
	if err != nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Unable to retrieve %s information", provider),
			Err:        err,
		}
	}

	switch provider {
	case portaineree.CloudProviderCivo:
		civoInfo, err := handler.cloudClusterInfoService.CivoGetInfo(credential, force)
		if err != nil {
			return &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to retrieve Civo information",
				Err:        err,
			}
		}

		return response.JSON(w, civoInfo)

	case portaineree.CloudProviderDigitalOcean:

		digitalOceanInfo, err := handler.cloudClusterInfoService.DigitalOceanGetInfo(credential, force)
		if err != nil {
			return &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to retrieve DigitalOcean information",
				Err:        err,
			}
		}

		return response.JSON(w, digitalOceanInfo)

	case portaineree.CloudProviderLinode:

		linodeInfo, err := handler.cloudClusterInfoService.LinodeGetInfo(credential, force)
		if err != nil {
			return &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to retrieve Linode information",
				Err:        err,
			}
		}

		return response.JSON(w, linodeInfo)

	case portaineree.CloudProviderGKE:
		gkeInfo, err := handler.cloudClusterInfoService.GKEGetInfo(credential, force)
		if err != nil {
			return &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to retrieve GKE information",
				Err:        err,
			}
		}

		return response.JSON(w, gkeInfo)

	case portaineree.CloudProviderAzure:

		azureInfo, err := handler.cloudClusterInfoService.AzureGetInfo(credential, force)
		if err != nil {
			return &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to retrieve Azure information",
				Err:        err,
			}
		}

		return response.JSON(w, azureInfo)

	case portaineree.CloudProviderAmazon:
		awsInfo, err := handler.cloudClusterInfoService.AmazonEksGetInfo(credential, force)
		if err != nil {
			return &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to retrieve AWS information",
				Err:        err,
			}
		}

		return response.JSON(w, awsInfo)
	}

	return &httperror.HandlerError{
		StatusCode: http.StatusInternalServerError,
		Message:    "Unable to get Kaas provider info",
		Err:        errors.New("invalid provider route parameter. Valid values: civo, digitalocean, linode, azure, gke"),
	}
}
