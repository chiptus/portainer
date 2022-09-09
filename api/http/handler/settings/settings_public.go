package settings

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type publicSettingsResponse struct {
	// URL to a logo that will be displayed on the login page as well as on top of the sidebar. Will use default Portainer logo when value is empty string
	LogoURL string `json:"LogoURL" example:"https://mycompany.mydomain.tld/logo.png"`
	// The content in plaintext used to display in the login page. Will hide when value is empty string
	CustomLoginBanner string `json:"CustomLoginBanner" example:"notice or agreement"`
	// Active authentication method for the Portainer instance. Valid values are: 1 for internal, 2 for LDAP, or 3 for oauth
	AuthenticationMethod portaineree.AuthenticationMethod `json:"AuthenticationMethod" example:"1"`
	// The minimum required length for a password of any user when using internal auth mode
	RequiredPasswordLength int `json:"RequiredPasswordLength" example:"1"`
	// Whether edge compute features are enabled
	EnableEdgeComputeFeatures bool `json:"EnableEdgeComputeFeatures" example:"true"`
	// Supported feature flags
	Features map[portaineree.Feature]bool `json:"Features"`
	// The URL used for oauth login
	OAuthLoginURI string `json:"OAuthLoginURI" example:"https://gitlab.com/oauth"`
	// The URL used for oauth logout
	OAuthLogoutURI string `json:"OAuthLogoutURI" example:"https://gitlab.com/oauth/logout"`
	// Whether portainer internal auth view will be hidden
	OAuthHideInternalAuth bool `json:"OAuthHideInternalAuth" example:"true"`
	// Whether telemetry is enabled
	EnableTelemetry bool `json:"EnableTelemetry" example:"true"`
	// The expiry of a Kubeconfig
	KubeconfigExpiry string `example:"24h" default:"0"`
	// Whether team sync is enabled
	TeamSync bool `json:"TeamSync" example:"true"`

	DefaultRegistry struct {
		Hide bool `json:"Hide" example:"false"`
	}

	Edge struct {
		// Whether the device has been started in edge async mode
		AsyncMode bool
		// The ping interval for edge agent - used in edge async mode [seconds]
		PingInterval int `json:"PingInterval" example:"60"`
		// The snapshot interval for edge agent - used in edge async mode [seconds]
		SnapshotInterval int `json:"SnapshotInterval" example:"60"`
		// The command list interval for edge agent - used in edge async mode [seconds]
		CommandInterval int `json:"CommandInterval" example:"60"`
		// The check in interval for edge agent (in seconds) - used in non async mode [seconds]
		CheckinInterval int `example:"60"`
	}
}

// @id SettingsPublic
// @summary Retrieve Portainer public settings
// @description Retrieve public settings. Returns a small set of settings that are not reserved to administrators only.
// @description **Access policy**: public
// @tags settings
// @produce json
// @success 200 {object} publicSettingsResponse "Success"
// @failure 500 "Server error"
// @router /settings/public [get]
func (handler *Handler) settingsPublic(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve the settings from the database", Err: err}
	}

	publicSettings := generatePublicSettings(settings)
	return response.JSON(w, publicSettings)
}

func generatePublicSettings(appSettings *portaineree.Settings) *publicSettingsResponse {
	publicSettings := &publicSettingsResponse{
		LogoURL:                   appSettings.LogoURL,
		CustomLoginBanner:         appSettings.CustomLoginBanner,
		AuthenticationMethod:      appSettings.AuthenticationMethod,
		RequiredPasswordLength:    appSettings.InternalAuthSettings.RequiredPasswordLength,
		EnableEdgeComputeFeatures: appSettings.EnableEdgeComputeFeatures,
		EnableTelemetry:           appSettings.EnableTelemetry,
		KubeconfigExpiry:          appSettings.KubeconfigExpiry,
		Features:                  appSettings.FeatureFlagSettings,
		DefaultRegistry:           appSettings.DefaultRegistry,
	}

	publicSettings.Edge.AsyncMode = appSettings.Edge.AsyncMode
	publicSettings.Edge.PingInterval = appSettings.Edge.PingInterval
	publicSettings.Edge.SnapshotInterval = appSettings.Edge.SnapshotInterval
	publicSettings.Edge.CommandInterval = appSettings.Edge.CommandInterval
	publicSettings.Edge.CheckinInterval = appSettings.EdgeAgentCheckinInterval

	//if OAuth authentication is on, compose the related fields from application settings
	if publicSettings.AuthenticationMethod == portaineree.AuthenticationOAuth {
		publicSettings.OAuthLogoutURI = appSettings.OAuthSettings.LogoutURI
		publicSettings.OAuthHideInternalAuth = appSettings.OAuthSettings.HideInternalAuth
		publicSettings.OAuthLoginURI = fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s",
			appSettings.OAuthSettings.AuthorizationURI,
			appSettings.OAuthSettings.ClientID,
			appSettings.OAuthSettings.RedirectURI,
			appSettings.OAuthSettings.Scopes)
		publicSettings.TeamSync = appSettings.OAuthSettings.OAuthAutoMapTeamMemberships
		//control prompt=login param according to the SSO setting
		if !appSettings.OAuthSettings.SSO {
			publicSettings.OAuthLoginURI += "&prompt=login"
		}
	}
	//if LDAP authentication is on, compose the related fields from application settings
	if publicSettings.AuthenticationMethod == portaineree.AuthenticationLDAP && appSettings.LDAPSettings.GroupSearchSettings != nil {
		if len(appSettings.LDAPSettings.GroupSearchSettings) > 0 {
			publicSettings.TeamSync = len(appSettings.LDAPSettings.GroupSearchSettings[0].GroupBaseDN) > 0
		}
	}
	return publicSettings
}
