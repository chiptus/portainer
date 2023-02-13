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
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	force, err := request.RetrieveBooleanQueryParameter(r, "force", true)
	if err != nil {
		return httperror.BadRequest("Failed parsing \"force\" boolean", err)
	}

	credentialId, _ := request.RetrieveNumericQueryParameter(r, "credentialId", true)
	if credentialId == 0 {
		return httperror.InternalServerError("Missing credential id in the query parameter", err)
	}

	credential, err := handler.dataStore.CloudCredential().GetByID(models.CloudCredentialID(credentialId))
	if err != nil {
		return httperror.InternalServerError(fmt.Sprintf("Unable to retrieve %s information", provider), err)
	}

	switch provider {
	case portaineree.CloudProviderMicrok8s:
		nodeIP, _ := request.RetrieveQueryParameter(r, "nodeip", true)
		if nodeIP == "" {
			return httperror.BadRequest("Missing required nodeip parameter", errors.New(request.ErrInvalidQueryParameter))
		}
		microK8sInfo, err := handler.cloudClusterInfoService.Microk8sGetAddons(credential, nodeIP)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve MicroK8s information", err)
		}

		return response.JSON(w, microK8sInfo)

	case portaineree.CloudProviderCivo:
		civoInfo, err := handler.cloudClusterInfoService.CivoGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Civo information", err)
		}

		return response.JSON(w, civoInfo)

	case portaineree.CloudProviderDigitalOcean:

		digitalOceanInfo, err := handler.cloudClusterInfoService.DigitalOceanGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve DigitalOcean information", err)
		}

		return response.JSON(w, digitalOceanInfo)

	case portaineree.CloudProviderLinode:

		linodeInfo, err := handler.cloudClusterInfoService.LinodeGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Linode information", err)
		}

		return response.JSON(w, linodeInfo)

	case portaineree.CloudProviderGKE:
		gkeInfo, err := handler.cloudClusterInfoService.GKEGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve GKE information", err)
		}

		return response.JSON(w, gkeInfo)

	case portaineree.CloudProviderAzure:

		azureInfo, err := handler.cloudClusterInfoService.AzureGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Azure information", err)
		}

		return response.JSON(w, azureInfo)

	case portaineree.CloudProviderAmazon:
		awsInfo, err := handler.cloudClusterInfoService.AmazonEksGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve AWS information", err)
		}

		return response.JSON(w, awsInfo)
	}

	return httperror.InternalServerError("Unable to get Kaas provider info", errors.New("invalid provider route parameter. Valid values: civo, digitalocean, linode, azure, gke"))
}
