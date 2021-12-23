package status

import (
	"encoding/json"
	"net/http"

	"github.com/coreos/go-semver/semver"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/client"

	"github.com/portainer/libhttp/response"
)

type inspectVersionResponse struct {
	// Whether portainer has an update available
	UpdateAvailable bool `json:"UpdateAvailable" example:"false"`
	// The latest version available
	LatestVersion string `json:"LatestVersion" example:"2.0.0"`
}

type githubData struct {
	TagName string `json:"tag_name"`
}

// @id StatusInspectVersion
// @summary Check for portainer updates
// @description Check if portainer has an update available
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags status
// @produce json
// @success 200 {object} inspectVersionResponse "Success"
// @router /status/version [get]
func (handler *Handler) statusInspectVersion(w http.ResponseWriter, r *http.Request) {
	motd, err := client.Get(portaineree.VersionCheckURL, 5)
	if err != nil {
		response.JSON(w, &inspectVersionResponse{UpdateAvailable: false})
		return
	}

	var data githubData
	err = json.Unmarshal(motd, &data)
	if err != nil {
		response.JSON(w, &inspectVersionResponse{UpdateAvailable: false})
		return
	}

	resp := inspectVersionResponse{
		UpdateAvailable: false,
	}

	currentVersion := semver.New(portaineree.APIVersion)
	latestVersion := semver.New(data.TagName)
	if currentVersion.LessThan(*latestVersion) {
		// disable update version available for EE
		resp.UpdateAvailable = false
		resp.LatestVersion = data.TagName
	}

	response.JSON(w, &resp)
}
