package settings

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/httpclient"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/pkg/libhelm"
)

type mTLSPayload struct {
	UseSeparateCert *bool
	CaCert          *string
	Cert            *string
	Key             *string
}

type settingsUpdatePayload struct {
	// URL to a logo that will be displayed on the login page as well as on top of the sidebar. Will use default Portainer logo when value is empty string
	LogoURL *string `example:"https://mycompany.mydomain.tld/logo.png"`
	// The content in plaintext used to display in the login page. Will hide when value is empty string
	CustomLoginBanner *string `example:"notice or agreement"`
	// A list of label name & value that will be used to hide containers when querying containers
	BlackListedLabels []portaineree.Pair
	// Active authentication method for the Portainer instance. Valid values are: 1 for internal, 2 for LDAP, or 3 for oauth
	AuthenticationMethod *int `example:"1"`
	InternalAuthSettings *portaineree.InternalAuthSettings
	LDAPSettings         *portaineree.LDAPSettings
	OAuthSettings        *portaineree.OAuthSettings
	// The interval in which environment(endpoint) snapshots are created
	SnapshotInterval *string `example:"5m"`
	// URL to the templates that will be displayed in the UI when navigating to App Templates
	TemplatesURL *string `example:"https://raw.githubusercontent.com/portainer/templates/master/templates.json"`
	// Deployment options for encouraging deployment as code
	GlobalDeploymentOptions *portaineree.GlobalDeploymentOptions
	// Show the Kompose build option (discontinued in 2.18)
	ShowKomposeBuildOption *bool `json:"ShowKomposeBuildOption" example:"false"`
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

	Edge struct {
		// The ping interval for edge agent - used in edge async mode (in seconds)
		PingInterval *int `json:"PingInterval" example:"5"`
		// The snapshot interval for edge agent - used in edge async mode (in seconds)
		SnapshotInterval *int `json:"SnapshotInterval" example:"5"`
		// The command list interval for edge agent - used in edge async mode (in seconds)
		CommandInterval *int `json:"CommandInterval" example:"5"`
		// AsyncMode enables edge agent to run in async mode by default
		AsyncMode *bool

		MTLS mTLSPayload
		// The address where the tunneling server can be reached by Edge agents
		TunnelServerAddress *string
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
		return httperror.BadRequest("Invalid request payload", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the settings from the database", err)
	}

	if handler.demoService.IsDemo() {
		payload.EnableTelemetry = nil
		payload.LogoURL = nil
	}

	if payload.AuthenticationMethod != nil {
		settings.AuthenticationMethod = portaineree.AuthenticationMethod(*payload.AuthenticationMethod)
	}

	if payload.LogoURL != nil {
		settings.LogoURL = *payload.LogoURL
	}

	if payload.CustomLoginBanner != nil {
		settings.CustomLoginBanner = *payload.CustomLoginBanner
	}

	if payload.TemplatesURL != nil {
		settings.TemplatesURL = *payload.TemplatesURL
	}

	// update the global deployment options, and the environment deployment options if they have changed
	if payload.GlobalDeploymentOptions != nil {
		oldPerEnvOverride := settings.GlobalDeploymentOptions.PerEnvOverride
		newPerEnvOverride := payload.GlobalDeploymentOptions.PerEnvOverride

		if oldPerEnvOverride != newPerEnvOverride {
			httpErr := handler.updateEnvironmentDeploymentType()
			if httpErr != nil {
				return httpErr
			}
		}

		settings.GlobalDeploymentOptions = *payload.GlobalDeploymentOptions
	}

	if payload.ShowKomposeBuildOption != nil {
		settings.ShowKomposeBuildOption = *payload.ShowKomposeBuildOption
	}

	if payload.HelmRepositoryURL != nil {
		if *payload.HelmRepositoryURL != "" {

			newHelmRepo := strings.TrimSuffix(strings.ToLower(*payload.HelmRepositoryURL), "/")

			if newHelmRepo != settings.HelmRepositoryURL && newHelmRepo != portaineree.DefaultHelmRepositoryURL {
				client := httpclient.NewWithOptions(
					httpclient.WithClientCertificate(handler.FileService.GetSSLClientCertPath()),
				)

				err := libhelm.ValidateHelmRepositoryURL(*payload.HelmRepositoryURL, client)
				if err != nil {
					return httperror.BadRequest("Invalid Helm repository URL. Must correspond to a valid URL format", err)
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

	if payload.InternalAuthSettings != nil {
		settings.InternalAuthSettings.RequiredPasswordLength = payload.InternalAuthSettings.RequiredPasswordLength
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
				return httperror.BadRequest("oauth claim name required", errors.New("provided oauth team membership claim name is empty"))
			}

			for _, mapping := range payload.OAuthSettings.TeamMemberships.OAuthClaimMappings {
				if mapping.ClaimValRegex == "" || mapping.Team == 0 {
					return httperror.BadRequest("invalid oauth mapping provided", fmt.Errorf("invalid oauth team membership mapping; mapping=%v", mapping))
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
			return httperror.InternalServerError("Unable to update snapshot interval", err)
		}
	}

	if payload.EdgeAgentCheckinInterval != nil {
		settings.EdgeAgentCheckinInterval = *payload.EdgeAgentCheckinInterval
	}

	if payload.Edge.PingInterval != nil {
		settings.Edge.PingInterval = *payload.Edge.PingInterval
	}

	if payload.Edge.SnapshotInterval != nil {
		settings.Edge.SnapshotInterval = *payload.Edge.SnapshotInterval
	}

	if payload.Edge.CommandInterval != nil {
		settings.Edge.CommandInterval = *payload.Edge.CommandInterval
	}

	if payload.Edge.TunnelServerAddress != nil {
		settings.Edge.TunnelServerAddress = *payload.Edge.TunnelServerAddress
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

	if payload.Edge.MTLS.UseSeparateCert != nil {
		settings.Edge.MTLS.UseSeparateCert = *payload.Edge.MTLS.UseSeparateCert

		if *payload.Edge.MTLS.UseSeparateCert {
			// If mtls is enabled, but cert and key are not provided
			// we should check if there is an existing one in the database
			// If there is, we skip saving the empty cert and key provided in payload
			// If there isn't, we should return an error
			if payload.Edge.MTLS.Cert == nil || payload.Edge.MTLS.Key == nil || payload.Edge.MTLS.CaCert == nil {
				// Check if there is an existing cert and key in the database
				if settings.Edge.MTLS.CaCertFile == "" || settings.Edge.MTLS.CertFile == "" || settings.Edge.MTLS.KeyFile == "" {
					// no existing cert and key in the database
					return httperror.BadRequest("Unable to save mTLS settings", errors.New("mTLS settings are incomplete"))
				}
			} else {
				certPath, caCertPath, keyPath, err := handler.FileService.StoreMTLSCertificates(
					[]byte(*payload.Edge.MTLS.Cert),
					[]byte(*payload.Edge.MTLS.CaCert),
					[]byte(*payload.Edge.MTLS.Key))
				if err != nil {
					return httperror.InternalServerError("Unable to persist mTLS certificates", err)
				}

				settings.Edge.MTLS.CertFile = certPath
				settings.Edge.MTLS.CaCertFile = caCertPath
				settings.Edge.MTLS.KeyFile = keyPath
				err = handler.SSLService.SetMTLSCertificates([]byte(*payload.Edge.MTLS.CaCert), []byte(*payload.Edge.MTLS.Cert), []byte(*payload.Edge.MTLS.Key))
				if err != nil {
					return httperror.InternalServerError("Unable to set mtls certificates", err)
				}
			}
		} else {
			// If mtls is disabled, we should clear the cert and key
			settings.Edge.MTLS.CertFile = ""
			settings.Edge.MTLS.KeyFile = ""
			settings.Edge.MTLS.CaCertFile = ""
			handler.SSLService.DisableMTLS()
		}
	}

	err = handler.DataStore.Settings().UpdateSettings(settings)
	if err != nil {
		return httperror.InternalServerError("Unable to persist settings changes inside the database", err)
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
			return httperror.InternalServerError("Unable to remove TLS files from disk", err)
		}
	}
	return nil
}

// updateEnvironmentDeploymentType updates environment deployment options to nil when per env override global settings change
func (handler *Handler) updateEnvironmentDeploymentType() *httperror.HandlerError {
	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the environments from the database", err)
	}

	for _, endpoint := range endpoints {
		if endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment || endpoint.Type == portaineree.KubeConfigEnvironment || endpoint.Type == portaineree.KubernetesLocalEnvironment {
			// save database writes by only updating the envs that have deployment option values
			if endpoint.DeploymentOptions != nil {
				endpoint.DeploymentOptions = nil
				err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, &endpoint)
				if err != nil {
					return httperror.InternalServerError("Unable to update the deployment options for the environment", err)
				}
			}
		}
	}
	return nil
}
