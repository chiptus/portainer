package auth

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestIsLDAPAdmin_Match(t *testing.T) {

	ldapService := testhelpers.NewLDAPService()

	mockLDAPSettings := &portaineree.LDAPSettings{
		AdminGroups: []string{"manager", "stuff"},
	}

	isLDAPAdmin, err := isLDAPAdmin("username", ldapService, mockLDAPSettings)
	assert.NoError(t, err)
	assert.Equal(t, true, isLDAPAdmin)
}

func TestIsLDAPAdmin_NotMatch(t *testing.T) {
	ldapService := testhelpers.NewLDAPService()

	mockLDAPSettings := &portaineree.LDAPSettings{
		AdminGroups: []string{"admin", "manager"},
	}

	isLDAPAdmin, err := isLDAPAdmin("username", ldapService, mockLDAPSettings)
	assert.NoError(t, err)
	assert.Equal(t, false, isLDAPAdmin)
}
