package helm

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/portainer/portainer/pkg/libhelm"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/httpclient"
)

type helmUserRepositoryResponse struct {
	GlobalRepository string                           `json:"GlobalRepository"`
	UserRepositories []portaineree.HelmUserRepository `json:"UserRepositories"`
}

type addHelmRepoUrlPayload struct {
	URL string `json:"url"`

	clientCert string
}

func (p *addHelmRepoUrlPayload) Validate(_ *http.Request) error {
	client := httpclient.NewWithOptions(
		httpclient.WithClientCertificate(p.clientCert),
	)
	return libhelm.ValidateHelmRepositoryURL(p.URL, client)
}

// @id HelmUserRepositoryCreate
// @summary Create a user helm repository
// @description Create a user helm repository.
// @description **Access policy**: authenticated
// @tags helm
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param payload body addHelmRepoUrlPayload true "Helm Repository"
// @success 200 {object} portaineree.HelmUserRepository "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @router /endpoints/{id}/kubernetes/helm/repositories [post]
func (handler *Handler) userCreateHelmRepo(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}
	userID := portaineree.UserID(tokenData.ID)

	httperr := handler.authoriseHelmOperation(r, portaineree.OperationHelmRepoCreate)
	if httperr != nil {
		return httperr
	}

	p := new(addHelmRepoUrlPayload)
	p.clientCert = handler.fileService.GetSSLClientCertPath()
	err = request.DecodeAndValidateJSONPayload(r, p)
	if err != nil {
		return httperror.BadRequest("Invalid Helm repository URL", err)
	}

	// lowercase, remove trailing slash
	p.URL = strings.TrimSuffix(strings.ToLower(p.URL), "/")

	records, err := handler.dataStore.HelmUserRepository().HelmUserRepositoryByUserID(userID)
	if err != nil {
		return httperror.InternalServerError("Unable to access the DataStore", err)
	}

	// check if repo already exists - by doing case insensitive comparison
	for _, record := range records {
		if strings.EqualFold(record.URL, p.URL) {
			errMsg := "Helm repo already registered for user"
			return httperror.BadRequest(errMsg, errors.New(errMsg))
		}
	}

	record := portaineree.HelmUserRepository{
		UserID: userID,
		URL:    p.URL,
	}

	err = handler.dataStore.HelmUserRepository().Create(&record)
	if err != nil {
		return httperror.InternalServerError("Unable to save a user Helm repository URL", err)
	}

	return response.JSON(w, record)
}

// @id HelmUserRepositoriesList
// @summary List a users helm repositories
// @description Inspect a user helm repositories.
// @description **Access policy**: authenticated
// @tags helm
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} helmUserRepositoryResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @router /endpoints/{id}/kubernetes/helm/repositories [get]
func (handler *Handler) userGetHelmRepos(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}
	userID := portaineree.UserID(tokenData.ID)

	httperr := handler.authoriseHelmOperation(r, portaineree.OperationHelmRepoList)
	if httperr != nil {
		return httperr
	}

	settings, err := handler.dataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	userRepos, err := handler.dataStore.HelmUserRepository().HelmUserRepositoryByUserID(userID)
	if err != nil {
		return httperror.InternalServerError("Unable to get user Helm repositories", err)
	}

	resp := helmUserRepositoryResponse{
		GlobalRepository: settings.HelmRepositoryURL,
		UserRepositories: userRepos,
	}

	return response.JSON(w, resp)
}
