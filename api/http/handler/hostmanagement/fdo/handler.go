package fdo

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router
	DataStore   dataservices.DataStore
	FileService portainer.FileService
}

func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore, fileService portainer.FileService) *Handler {
	h := &Handler{
		Router:      mux.NewRouter(),
		DataStore:   dataStore,
		FileService: fileService,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess)

	adminRouter.Handle("/fdo/configure", httperror.LoggerHandler(h.fdoConfigure)).Methods(http.MethodPost)
	adminRouter.Handle("/fdo/list", httperror.LoggerHandler(h.fdoListAll)).Methods(http.MethodGet)
	adminRouter.Handle("/fdo/register", httperror.LoggerHandler(h.fdoRegisterDevice)).Methods(http.MethodPost)
	adminRouter.Handle("/fdo/configure/{guid}", httperror.LoggerHandler(h.fdoConfigureDevice)).Methods(http.MethodPost)
	adminRouter.Handle("/fdo/profiles", httperror.LoggerHandler(h.fdoProfileList)).Methods(http.MethodGet)
	adminRouter.Handle("/fdo/profiles", httperror.LoggerHandler(h.createProfile)).Methods(http.MethodPost)
	adminRouter.Handle("/fdo/profiles/{id}", httperror.LoggerHandler(h.fdoProfileInspect)).Methods(http.MethodGet)
	adminRouter.Handle("/fdo/profiles/{id}", httperror.LoggerHandler(h.updateProfile)).Methods(http.MethodPut)
	adminRouter.Handle("/fdo/profiles/{id}", httperror.LoggerHandler(h.deleteProfile)).Methods(http.MethodDelete)
	adminRouter.Handle("/fdo/profiles/{id}/duplicate", httperror.LoggerHandler(h.duplicateProfile)).Methods(http.MethodPost)

	return h
}
