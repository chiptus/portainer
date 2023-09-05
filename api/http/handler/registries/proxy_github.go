package registries

import (
	"fmt"
	"net/http"

	"github.com/portainer/portainer-ee/api/github/packages"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type catalogPayload struct {
	Repositories []string `json:"repositories"`
}

// Mimics GET /v2/_catalog for ghcr.io
func proxyGithubRegistriesCatalog(ghPackages *packages.Packages) httperror.LoggerHandler {
	return func(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
		packages, err := ghPackages.GetPackages()
		if err != nil {
			return httperror.InternalServerError(err.Error(), err)
		}

		payload := catalogPayload{
			Repositories: make([]string, len(packages)),
		}
		for i := range packages {
			payload.Repositories[i] = fmt.Sprintf("%s/%s", packages[i].Owner.Login, packages[i].Name)
		}

		return response.JSON(w, payload)
	}
}

// Mimics DELETE /v2/{userName}/{packageName}/manifests/<reference> for ghcr.io
func proxyGithubRegistriesDeleteManifest(ghPackages *packages.Packages) httperror.LoggerHandler {
	return func(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
		packageName, err := request.RetrieveRouteVariableValue(r, "packageName")
		if err != nil {
			return httperror.BadRequest("Invalid package name route variable", err)
		}

		digest, err := request.RetrieveRouteVariableValue(r, "reference")
		if err != nil {
			return httperror.BadRequest("Invalid digest route variable", err)
		}

		err = ghPackages.DeleteManifest(packageName, digest)
		if err != nil {
			return httperror.InternalServerError("Unable to delete manifest", err)
		}

		return response.Empty(w)
	}
}
