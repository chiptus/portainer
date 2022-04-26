package settings

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/portainer/libhelm"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
)

type settingsUpdatePayload struct {
	// URL to a logo that will be displayed on the login page as well as on top of the sidebar. Will use default Portainer logo when value is empty string
	LogoURL *string `example:"https://mycompany.mydomain.tld/logo.png"`
	// A list of label name & value that will be used to hide containers when querying containers
	BlackListedLabels []portaineree.Pair
	// Active authentication method for the Portainer instance. Valid values are: 1 for internal, 2 for LDAP, or 3 for oauth
	AuthenticationMethod *int `example:"1"`
	LDAPSettings         *portaineree.LDAPSettings
	OAuthSettings        *portaineree.OAuthSettings
	// The interval in which environment(endpoint) snapshots are created
	SnapshotInterval *string `example:"5m"`
	// URL to the templates that will be displayed in the UI when navigating to App Templates
	TemplatesURL *string `example:"https://raw.githubusercontent.com/portainer/templates/master/templates.json"`
	// Whether edge compute features are enabled
	EnableEdgeComputeFeatures *bool `example:"true"`
	// The duration of a user session
	UserSessionTimeout *string `example:"5m"`
	// The expiry of a Kubeconfig
	KubeconfigExpiry *string `example:"24h" default:"0"`
	// Whether telemetry is enabled
	EnableTelemetry *bool `example:"false"`
	// Helm repository URL
	HelmRepositoryURL *string `example:"https://charts.bitnami.com/bitnami"`
	// Kubec	tl Shell Image Name/Tag
	KubectlShellImage *string `example:"portainer/kubectl-shell:latest"`
	// TrustOnFirstConnect makes Portainer accepting edge agent connection by default
	TrustOnFirstConnect *bool `example:"false"`
	// EnforceEdgeID makes Portainer store the Edge ID instead of accepting anyone
	EnforceEdgeID *bool `example:"false"`
	// EdgePortainerURL is the URL that is exposed to edge agents
	EdgePortainerURL *string `json:"EdgePortainerURL"`
	// The default check in interval for edge agent (in seconds)
	EdgeAgentCheckinInterval *int `example:"5"`
	// The command list interval for edge agent - used in edge async mode (in seconds)
	EdgeCommandInterval *int `json:"EdgeCommandInterval" example:"5"`
	// The ping interval for edge agent - used in edge async mode (in seconds)
	EdgePingInterval *int `json:"EdgePingInterval" example:"5"`
	// The snapshot interval for edge agent - used in edge async mode (in seconds)
	EdgeSnapshotInterval *int `json:"EdgeSnapshotInterval" example:"5"`

	CloudApiKeys struct {
		// CivoApiKey represents an authentication key for the Civo API
		CivoApiKey *string `example:"DgJ33kwIhnHumQFyc8ihGwWJql9cC8UJDiBhN8YImKqiX"`
		// DigitalOceanToken represents an authentication token for the DigitalOcean API
		DigitalOceanToken *string `example:"dop_v1_n9rq7dkcbg3zb3bewtk9nnvmfkyfnr94d42n28lym22vhqu23rtkllsldygqm22v"`
		// LinodeToken represents an authentication token for the Linode API
		LinodeToken *string `example:"92gsh9r9u5helgs4eibcuvlo403vm45hrmc6mzbslotnrqmkwc1ovqgmolcyq0wc"`
	}
}

func (payload *settingsUpdatePayload) Validate(r *http.Request) error {
	if payload.AuthenticationMethod != nil && *payload.AuthenticationMethod != 1 && *payload.AuthenticationMethod != 2 && *payload.AuthenticationMethod != 3 {
		return errors.New("Invalid authentication method value. Value must be one of: 1 (internal), 2 (LDAP/AD) or 3 (OAuth)")
	}
	if payload.LogoURL != nil && *payload.LogoURL != "" && !govalidator.IsURL(*payload.LogoURL) {
		return errors.New("Invalid logo URL. Must correspond to a valid URL format")
	}
	if payload.TemplatesURL != nil && *payload.TemplatesURL != "" && !govalidator.IsURL(*payload.TemplatesURL) {
		return errors.New("Invalid external templates URL. Must correspond to a valid URL format")
	}
	if payload.HelmRepositoryURL != nil && *payload.HelmRepositoryURL != "" && !govalidator.IsURL(*payload.HelmRepositoryURL) {
		return errors.New("Invalid Helm repository URL. Must correspond to a valid URL format")
	}
	if payload.UserSessionTimeout != nil {
		_, err := time.ParseDuration(*payload.UserSessionTimeout)
		if err != nil {
			return errors.New("Invalid user session timeout")
		}
	}
	if payload.KubeconfigExpiry != nil {
		_, err := time.ParseDuration(*payload.KubeconfigExpiry)
		if err != nil {
			return errors.New("Invalid Kubeconfig Expiry")
		}
	}

	if payload.AuthenticationMethod != nil && *payload.AuthenticationMethod == 2 {
		if payload.LDAPSettings == nil {
			return errors.New("Invalid LDAP Configuration")
		}
		if len(payload.LDAPSettings.URLs) == 0 {
			return errors.New("Invalid LDAP URLs. At least one URL is required")
		}
		if payload.LDAPSettings.AdminAutoPopulate && len(payload.LDAPSettings.AdminGroupSearchSettings) == 0 {
			return errors.New("Missing Admin group Search settings. when AdminAutoPopulate is true, at least one settings is required")
		}
		if !payload.LDAPSettings.AdminAutoPopulate && len(payload.LDAPSettings.AdminGroups) > 0 {
			payload.LDAPSettings.AdminGroups = []string{}
		}
	}

	if payload.EdgePortainerURL != nil {
		_, err := edge.ParseHostForEdge(*payload.EdgePortainerURL)
		if err != nil {
			return err
		}
	}

	return nil
}

// @id SettingsUpdate
// @summary Update Portainer settings
// @description Update Portainer settings.
// @description **Access policy**: administrator
// @tags settings
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body settingsUpdatePayload true "New settings"
// @success 200 {object} portaineree.Settings "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /settings [put]
func (handler *Handler) settingsUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload settingsUpdatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve the settings from the database", Err: err}
	}

	if payload.AuthenticationMethod != nil {
		settings.AuthenticationMethod = portaineree.AuthenticationMethod(*payload.AuthenticationMethod)
	}

	if payload.LogoURL != nil {
		settings.LogoURL = *payload.LogoURL
	}

	if payload.TemplatesURL != nil {
		settings.TemplatesURL = *payload.TemplatesURL
	}

	if payload.HelmRepositoryURL != nil {
		if *payload.HelmRepositoryURL != "" {

			newHelmRepo := strings.TrimSuffix(strings.ToLower(*payload.HelmRepositoryURL), "/")

			if newHelmRepo != settings.HelmRepositoryURL && newHelmRepo != portaineree.DefaultHelmRepositoryURL {
				err := libhelm.ValidateHelmRepositoryURL(*payload.HelmRepositoryURL)
				if err != nil {
					return &httperror.HandlerError{http.StatusBadRequest, "Invalid Helm repository URL. Must correspond to a valid URL format", err}
				}
			}
			settings.HelmRepositoryURL = newHelmRepo
		} else {
			settings.HelmRepositoryURL = ""
		}
	}

	if payload.BlackListedLabels != nil {
		settings.BlackListedLabels = payload.BlackListedLabels
	}

	if payload.LDAPSettings != nil {
		ldapReaderDN := settings.LDAPSettings.ReaderDN
		ldapPassword := settings.LDAPSettings.Password

		if payload.LDAPSettings.ReaderDN != "" {
			ldapReaderDN = payload.LDAPSettings.ReaderDN
		}

		if payload.LDAPSettings.Password != "" {
			ldapPassword = payload.LDAPSettings.Password
		}

		if payload.LDAPSettings.AnonymousMode {
			ldapReaderDN = ""
			ldapPassword = ""
		}

		settings.LDAPSettings = *payload.LDAPSettings
		settings.LDAPSettings.ReaderDN = ldapReaderDN
		settings.LDAPSettings.Password = ldapPassword
	}

	if payload.OAuthSettings != nil {
		clientSecret := payload.OAuthSettings.ClientSecret
		if clientSecret == "" {
			clientSecret = settings.OAuthSettings.ClientSecret
		}
		kubeSecret := payload.OAuthSettings.KubeSecretKey
		if kubeSecret == nil {
			kubeSecret = settings.OAuthSettings.KubeSecretKey
		}
		//if SSO is switched off, then make sure HideInternalAuth is switched off
		if !payload.OAuthSettings.SSO && payload.OAuthSettings.HideInternalAuth {
			payload.OAuthSettings.HideInternalAuth = false
		}
		settings.OAuthSettings = *payload.OAuthSettings
		settings.OAuthSettings.ClientSecret = clientSecret
		settings.OAuthSettings.KubeSecretKey = kubeSecret

		if !payload.OAuthSettings.OAuthAutoMapTeamMemberships || payload.OAuthSettings.TeamMemberships.OAuthClaimName == "" {
			settings.OAuthSettings.TeamMemberships.OAuthClaimMappings = []portaineree.OAuthClaimMappings{}
		}

		if payload.OAuthSettings.OAuthAutoMapTeamMemberships {
			// throw errors on invalid values
			if payload.OAuthSettings.TeamMemberships.OAuthClaimName == "" {
				return &httperror.HandlerError{
					StatusCode: http.StatusBadRequest,
					Message:    "oauth claim name required",
					Err:        errors.New("provided oauth team membership claim name is empty"),
				}
			}

			for _, mapping := range payload.OAuthSettings.TeamMemberships.OAuthClaimMappings {
				if mapping.ClaimValRegex == "" || mapping.Team == 0 {
					return &httperror.HandlerError{
						StatusCode: http.StatusBadRequest,
						Message:    "invalid oauth mapping provided",
						Err:        fmt.Errorf("invalid oauth team membership mapping; mapping=%v", mapping),
					}
				}
			}
		} else {
			// clear out redundant values
			settings.OAuthSettings.TeamMemberships.OAuthClaimName = ""
			settings.OAuthSettings.TeamMemberships.OAuthClaimMappings = []portaineree.OAuthClaimMappings{}
		}

		if !payload.OAuthSettings.TeamMemberships.AdminAutoPopulate {
			settings.OAuthSettings.TeamMemberships.AdminGroupClaimsRegexList = []string{}
		}
	}

	if payload.EnableEdgeComputeFeatures != nil {
		settings.EnableEdgeComputeFeatures = *payload.EnableEdgeComputeFeatures
	}

	if payload.TrustOnFirstConnect != nil {
		settings.TrustOnFirstConnect = *payload.TrustOnFirstConnect
	}

	if payload.EnforceEdgeID != nil {
		settings.EnforceEdgeID = *payload.EnforceEdgeID
	}

	if payload.EdgePortainerURL != nil {
		settings.EdgePortainerURL = *payload.EdgePortainerURL
	}

	if payload.SnapshotInterval != nil && *payload.SnapshotInterval != settings.SnapshotInterval {
		err := handler.updateSnapshotInterval(settings, *payload.SnapshotInterval)
		if err != nil {
			return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to update snapshot interval", Err: err}
		}
	}

	if payload.EdgeAgentCheckinInterval != nil {
		settings.EdgeAgentCheckinInterval = *payload.EdgeAgentCheckinInterval
	}

	if payload.EdgeCommandInterval != nil {
		settings.EdgeCommandInterval = *payload.EdgeCommandInterval
	}

	if payload.EdgePingInterval != nil {
		settings.EdgePingInterval = *payload.EdgePingInterval
	}

	if payload.EdgeSnapshotInterval != nil {
		settings.EdgeSnapshotInterval = *payload.EdgeSnapshotInterval
	}

	if payload.KubeconfigExpiry != nil {
		settings.KubeconfigExpiry = *payload.KubeconfigExpiry
	}

	if payload.UserSessionTimeout != nil {
		settings.UserSessionTimeout = *payload.UserSessionTimeout

		userSessionDuration, _ := time.ParseDuration(*payload.UserSessionTimeout)

		handler.JWTService.SetUserSessionDuration(userSessionDuration)
	}

	if payload.EnableTelemetry != nil {
		settings.EnableTelemetry = *payload.EnableTelemetry
	}

	tlsError := handler.updateTLS(settings)
	if tlsError != nil {
		return tlsError
	}

	if payload.KubectlShellImage != nil {
		settings.KubectlShellImage = *payload.KubectlShellImage
	}

	if payload.CloudApiKeys.CivoApiKey != nil {
		settings.CloudApiKeys.CivoApiKey = *payload.CloudApiKeys.CivoApiKey
	}

	if payload.CloudApiKeys.DigitalOceanToken != nil {
		settings.CloudApiKeys.DigitalOceanToken = *payload.CloudApiKeys.DigitalOceanToken
	}

	if payload.CloudApiKeys.LinodeToken != nil {
		settings.CloudApiKeys.LinodeToken = *payload.CloudApiKeys.LinodeToken
	}

	err = handler.DataStore.Settings().UpdateSettings(settings)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist settings changes inside the database", Err: err}
	}

	hideFields(settings)
	return response.JSON(w, settings)
}

func (handler *Handler) updateSnapshotInterval(settings *portaineree.Settings, snapshotInterval string) error {
	settings.SnapshotInterval = snapshotInterval

	return handler.SnapshotService.SetSnapshotInterval(snapshotInterval)
}

func (handler *Handler) updateTLS(settings *portaineree.Settings) *httperror.HandlerError {
	if (settings.LDAPSettings.TLSConfig.TLS || settings.LDAPSettings.StartTLS) && !settings.LDAPSettings.TLSConfig.TLSSkipVerify {
		caCertPath, _ := handler.FileService.GetPathForTLSFile(filesystem.LDAPStorePath, portainer.TLSFileCA)
		settings.LDAPSettings.TLSConfig.TLSCACertPath = caCertPath
	} else {
		settings.LDAPSettings.TLSConfig.TLSCACertPath = ""
		err := handler.FileService.DeleteTLSFiles(filesystem.LDAPStorePath)
		if err != nil {
			return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to remove TLS files from disk", Err: err}
		}
	}
	return nil
}
