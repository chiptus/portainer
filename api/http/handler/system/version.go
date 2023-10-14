package system

import (
	"net/http"
	"os"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/build"
	portainer "github.com/portainer/portainer/api"
	system "github.com/portainer/portainer/api/http/handler/system"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/rs/zerolog/log"
)

type versionResponse struct {
	// Whether portainer has an update available
	UpdateAvailable bool `json:"UpdateAvailable" example:"false"`
	// The latest version available
	LatestVersion string `json:"LatestVersion" example:"2.0.0"`

	ServerVersion   string
	ServerEdition   string `json:"ServerEdition" example:"CE/EE"`
	DatabaseVersion string
	Build           BuildInfo
}

type BuildInfo struct {
	BuildNumber    string
	ImageTag       string
	NodejsVersion  string
	YarnVersion    string
	WebpackVersion string
	GoVersion      string
}

// @id systemVersion
// @summary Check for portainer updates
// @description Check if portainer has an update available
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags system
// @produce json
// @success 200 {object} versionResponse "Success"
// @router /system/version [get]
func (handler *Handler) version(w http.ResponseWriter, r *http.Request) {

	var dbVer, edition string
	vs := handler.dataStore.Version()
	v, err := vs.Version()
	if err == nil {
		dbVer = v.SchemaVersion
		edition = portainer.SoftwareEdition(v.Edition).GetEditionLabel()
	}

	result := &versionResponse{
		ServerVersion:   portaineree.APIVersion,
		ServerEdition:   edition,
		DatabaseVersion: dbVer,
		Build: BuildInfo{
			BuildNumber:    build.BuildNumber,
			ImageTag:       build.ImageTag,
			NodejsVersion:  build.NodejsVersion,
			YarnVersion:    build.YarnVersion,
			WebpackVersion: build.WebpackVersion,
			GoVersion:      build.GoVersion,
		},
	}

	latestVersion := system.GetLatestVersion()
	if os.Getenv("TEST_UPDATE_PORTAINER") != "" || system.HasNewerVersion(portaineree.APIVersion, latestVersion) {
		result.UpdateAvailable = true
		result.LatestVersion = latestVersion
	}

	response.JSON(w, &result)
}

// @id Version
// @summary Check for portainer updates
// @deprecated
// @description Deprecated: use the `/system/version` endpoint instead.
// @description Check if portainer has an update available
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags status
// @produce json
// @success 200 {object} versionResponse "Success"
// @router /status/version [get]
func (handler *Handler) versionDeprecated(w http.ResponseWriter, r *http.Request) {
	log.Warn().Msg("The /status/version endpoint is deprecated, please use the /system/version endpoint instead")

	handler.version(w, r)
}
