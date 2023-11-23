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

	authenticatedRouter := h.NewRoute().Subrouter()
	authenticatedRouter.Use(bouncer.AuthenticatedAccess)
	authenticatedRouter.Handle("/gitops/repo/refs", httperror.LoggerHandler(h.gitOperationRepoRefs)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/gitops/repo/files/search", httperror.LoggerHandler(h.gitOperationRepoFilesSearch)).Methods(http.MethodPost)
	authenticatedRouter.Handle("/gitops/repo/file/preview", httperror.LoggerHandler(h.gitOperationRepoFilePreview)).Methods(http.MethodPost)

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
