package backup

import (
	"context"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/adminmonitor"
	operations "github.com/portainer/portainer-ee/api/backup"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/demo"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/offlinegate"
	"github.com/portainer/portainer-ee/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is an http handler responsible for backup and restore portainer state
type Handler struct {
	*mux.Router
	backupScheduler   *operations.BackupScheduler
	bouncer           security.BouncerService
	dataStore         dataservices.DataStore
	userActivityStore portaineree.UserActivityStore
	gate              *offlinegate.OfflineGate
	filestorePath     string
	shutdownTrigger   context.CancelFunc
	adminMonitor      *adminmonitor.Monitor
}

// NewHandler creates an new instance of backup handler
func NewHandler(
	bouncer security.BouncerService,
	dataStore dataservices.DataStore,
	userActivityStore portaineree.UserActivityStore,
	gate *offlinegate.OfflineGate,
	filestorePath string,
	backupScheduler *operations.BackupScheduler,
	shutdownTrigger context.CancelFunc,
	adminMonitor *adminmonitor.Monitor,
	demoService *demo.Service,
) *Handler {
	h := &Handler{
		Router:            mux.NewRouter(),
		bouncer:           bouncer,
		backupScheduler:   backupScheduler,
		dataStore:         dataStore,
		userActivityStore: userActivityStore,
		gate:              gate,
		filestorePath:     filestorePath,
		shutdownTrigger:   shutdownTrigger,
		adminMonitor:      adminMonitor,
	}

	adminRouter := h.NewRoute().Subrouter()
	adminRouter.Use(bouncer.PureAdminAccess)
	adminRouter.Handle("/backup/s3/settings", httperror.LoggerHandler(h.backupSettingsFetch)).Methods(http.MethodGet)

	publicRouter := h.NewRoute().Subrouter()
	publicRouter.Use(bouncer.PublicAccess)
	publicRouter.Handle("/backup/s3/status", httperror.LoggerHandler(h.backupStatusFetch)).Methods(http.MethodGet)

	demoRouter := h.NewRoute().Subrouter()
	demoRouter.Use(middlewares.RestrictDemoEnv(demoService.IsDemo))

	demoAdmin := demoRouter.NewRoute().Subrouter()
	demoAdmin.Use(bouncer.PureAdminAccess)

	demoAdmin.Handle("/backup/s3/settings", httperror.LoggerHandler(h.updateSettings)).Methods(http.MethodPost)
	demoAdmin.Handle("/backup/s3/execute", httperror.LoggerHandler(h.backupToS3)).Methods(http.MethodPost)
	demoAdmin.Handle("/backup", httperror.LoggerHandler(h.backup)).Methods(http.MethodPost)

	demoPublic := demoRouter.NewRoute().Subrouter()
	demoPublic.Use(bouncer.PublicAccess)

	demoPublic.Handle("/backup/s3/restore", httperror.LoggerHandler(h.restoreFromS3)).Methods(http.MethodPost)
	demoPublic.Handle("/restore", httperror.LoggerHandler(h.restore)).Methods(http.MethodPost)

	return h
}
