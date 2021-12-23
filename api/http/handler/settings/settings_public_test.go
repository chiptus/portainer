package settings

import (
	"fmt"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
)

const (
	dummyOAuthClientID          = "1a2b3c4d"
	dummyOAuthScopes            = "scopes"
	dummyOAuthAuthenticationURI = "example.com/auth"
	dummyOAuthRedirectURI       = "example.com/redirect"
	dummyOAuthLogoutURI         = "example.com/logout"
)

var (
	dummyOAuthLoginURI string
	mockAppSettings    *portaineree.Settings
)

func setup() {
	dummyOAuthLoginURI = fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s",
		dummyOAuthAuthenticationURI,
		dummyOAuthClientID,
		dummyOAuthRedirectURI,
		dummyOAuthScopes)
	mockAppSettings = &portaineree.Settings{
		AuthenticationMethod: portaineree.AuthenticationOAuth,
		OAuthSettings: portaineree.OAuthSettings{
			AuthorizationURI: dummyOAuthAuthenticationURI,
			ClientID:         dummyOAuthClientID,
			Scopes:           dummyOAuthScopes,
			RedirectURI:      dummyOAuthRedirectURI,
			LogoutURI:        dummyOAuthLogoutURI,
		},
	}
}

func TestGeneratePublicSettingsWithSSO(t *testing.T) {
	setup()
	mockAppSettings.OAuthSettings.SSO = true
	publicSettings := generatePublicSettings(mockAppSettings)
	if publicSettings.AuthenticationMethod != portaineree.AuthenticationOAuth {
		t.Errorf("wrong AuthenticationMethod, want: %d, got: %d", portaineree.AuthenticationOAuth, publicSettings.AuthenticationMethod)
	}
	if publicSettings.OAuthHideInternalAuth {
		t.Errorf("wrong OAuthHideInternalAuth, should be switched off by default, but got: %t", publicSettings.OAuthHideInternalAuth)
	}
	if publicSettings.OAuthLoginURI != dummyOAuthLoginURI {
		t.Errorf("wrong OAuthLoginURI when SSO is switched on, want: %s, got: %s", dummyOAuthLoginURI, publicSettings.OAuthLoginURI)
	}
	if publicSettings.OAuthLogoutURI != dummyOAuthLogoutURI {
		t.Errorf("wrong OAuthLogoutURI, want: %s, got: %s", dummyOAuthLogoutURI, publicSettings.OAuthLogoutURI)
	}
}

func TestGeneratePublicSettingsWithoutSSO(t *testing.T) {
	setup()
	mockAppSettings.OAuthSettings.SSO = false
	publicSettings := generatePublicSettings(mockAppSettings)
	if publicSettings.AuthenticationMethod != portaineree.AuthenticationOAuth {
		t.Errorf("wrong AuthenticationMethod, want: %d, got: %d", portaineree.AuthenticationOAuth, publicSettings.AuthenticationMethod)
	}
	if publicSettings.OAuthHideInternalAuth {
		t.Errorf("wrong OAuthHideInternalAuth, should be switched off by default, but got: %t", publicSettings.OAuthHideInternalAuth)
	}
	expectedOAuthLoginURI := dummyOAuthLoginURI + "&prompt=login"
	if publicSettings.OAuthLoginURI != expectedOAuthLoginURI {
		t.Errorf("wrong OAuthLoginURI when SSO is switched off, want: %s, got: %s", expectedOAuthLoginURI, publicSettings.OAuthLoginURI)
	}
	if publicSettings.OAuthLogoutURI != dummyOAuthLogoutURI {
		t.Errorf("wrong OAuthLogoutURI, want: %s, got: %s", dummyOAuthLogoutURI, publicSettings.OAuthLogoutURI)
	}
}

func TestGeneratePublicSettingsWithHideInternalAuth(t *testing.T) {
	setup()
	mockAppSettings.OAuthSettings.HideInternalAuth = true
	publicSettings := generatePublicSettings(mockAppSettings)
	if publicSettings.AuthenticationMethod != portaineree.AuthenticationOAuth {
		t.Errorf("wrong AuthenticationMethod, want: %d, got: %d", portaineree.AuthenticationOAuth, publicSettings.AuthenticationMethod)
	}
	if !publicSettings.OAuthHideInternalAuth {
		t.Errorf("wrong OAuthHideInternalAuth, should be switched on, but got: %t", publicSettings.OAuthHideInternalAuth)
	}
	if publicSettings.OAuthLogoutURI != dummyOAuthLogoutURI {
		t.Errorf("wrong OAuthLogoutURI, want: %s, got: %s", dummyOAuthLogoutURI, publicSettings.OAuthLogoutURI)
	}
}
