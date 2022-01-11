package tests

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt"
	"github.com/stretchr/testify/assert"
)

type teamBuilder struct {
	t     *testing.T
	count int
	store *bolt.Store
}

func (b *teamBuilder) createNew(name string) *portaineree.Team {
	b.count++
	team := &portaineree.Team{
		ID:   portaineree.TeamID(b.count),
		Name: name,
	}

	err := b.store.Team().CreateTeam(team)
	assert.NoError(b.t, err)

	return team
}