package tests

import (
	"testing"

	"github.com/portainer/portainer-ee/api/datastore"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

type teamBuilder struct {
	t     *testing.T
	count int
	store *datastore.Store
}

func (b *teamBuilder) createNew(name string) *portainer.Team {
	b.count++
	team := &portainer.Team{
		ID:   portainer.TeamID(b.count),
		Name: name,
	}

	err := b.store.TeamService.Create(team)
	assert.NoError(b.t, err)

	return team
}
