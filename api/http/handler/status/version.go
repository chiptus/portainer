package status

import (
	"net/http"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/build"

	"github.com/portainer/libhttp/response"
)

type versionResponse struct {
	// Whether portainer has an update available
	UpdateAvailable bool `json:"UpdateAvailable" example:"false"`
	// The latest version available
	LatestVersion string `json:"LatestVersion" example:"2.0.0"`

	ServerVersion   string
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

// @id Version
// @summary Check for portainer updates
// @description Check if portainer has an update available
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags status
// @produce json
// @success 200 {object} versionResponse "Success"
// @router /status/version [get]
func (handler *Handler) version(w http.ResponseWriter, r *http.Request) {
	result := &versionResponse{
		ServerVersion:   portaineree.APIVersion,
		DatabaseVersion: strconv.Itoa(portaineree.DBVersion),
		Build: BuildInfo{
			BuildNumber:    build.BuildNumber,
			ImageTag:       build.ImageTag,
			NodejsVersion:  build.NodejsVersion,
			YarnVersion:    build.YarnVersion,
			WebpackVersion: build.WebpackVersion,
			GoVersion:      build.GoVersion,
		},
	}

	response.JSON(w, &result)
}
