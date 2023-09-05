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

	demoRestrictedRouter := h.NewRoute().Subrouter()
	demoRestrictedRouter.Use(middlewares.RestrictDemoEnv(demoService.IsDemo))

	h.Handle("/backup/s3/settings", bouncer.RestrictedAccess(adminAccess(httperror.LoggerHandler(h.backupSettingsFetch)))).Methods(http.MethodGet)
	demoRestrictedRouter.Handle("/backup/s3/settings", bouncer.RestrictedAccess(adminAccess(httperror.LoggerHandler(h.updateSettings)))).Methods(http.MethodPost)
	h.Handle("/backup/s3/status", bouncer.PublicAccess(httperror.LoggerHandler(h.backupStatusFetch))).Methods(http.MethodGet)
	demoRestrictedRouter.Handle("/backup/s3/execute", bouncer.RestrictedAccess(adminAccess(httperror.LoggerHandler(h.backupToS3)))).Methods(http.MethodPost)
	demoRestrictedRouter.Handle("/backup/s3/restore", bouncer.PublicAccess(httperror.LoggerHandler(h.restoreFromS3))).Methods(http.MethodPost)
	demoRestrictedRouter.Handle("/backup", bouncer.RestrictedAccess(adminAccess(httperror.LoggerHandler(h.backup)))).Methods(http.MethodPost)
	demoRestrictedRouter.Handle("/restore", bouncer.PublicAccess(httperror.LoggerHandler(h.restore))).Methods(http.MethodPost)

	return h
}

func adminAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		securityContext, err := security.RetrieveRestrictedRequestContext(r)
		if err != nil {
			httperror.WriteError(w, http.StatusInternalServerError, "Unable to retrieve user info from request context", err)
			return
		}

		if !securityContext.IsAdmin {
			httperror.WriteError(w, http.StatusUnauthorized, "User is not authorized to perform the action", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
