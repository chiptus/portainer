package gitops

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle git repo operation
type Handler struct {
	*mux.Router
	dataStore   dataservices.DataStore
	GitService  portainer.GitService
	FileService portainer.FileService
}

func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore, gitService portainer.GitService, fileService portainer.FileService) *Handler {
	h := &Handler{
		Router:      mux.NewRouter(),
		dataStore:   dataStore,
		GitService:  gitService,
		FileService: fileService,
	}

	h.Handle("/gitops/repo/refs",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.gitOperationRepoRefs))).Methods(http.MethodPost)
	h.Handle("/gitops/repo/files/search",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.gitOperationRepoFilesSearch))).Methods(http.MethodPost)
	h.Handle("/gitops/repo/file/preview",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.gitOperationRepoFilePreview))).Methods(http.MethodPost)

	return h
}

func (handler *Handler) extractGitCredential(username, password string, credentialID int) (string, string, *httperror.HandlerError) {
	repositoryUsername := ""
	repositoryPassword := ""
	if credentialID != 0 {
		credential, err := handler.dataStore.GitCredential().Read(portaineree.GitCredentialID(credentialID))
		if err != nil {
			return "", "", httperror.InternalServerError("git credential not found", err)
		}

		repositoryUsername = credential.Username
		repositoryPassword = credential.Password
	}

	if password != "" {
		repositoryUsername = username
		repositoryPassword = password
	}
	return repositoryUsername, repositoryPassword, nil
}
