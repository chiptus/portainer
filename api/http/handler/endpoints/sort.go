package endpoints

import (
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/natsort"
)

type EndpointsByName []portaineree.Endpoint

func (e EndpointsByName) Len() int {
	return len(e)
}

func (e EndpointsByName) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e EndpointsByName) Less(i, j int) bool {
	return natsort.Compare(strings.ToLower(e[i].Name), strings.ToLower(e[j].Name))
}

type EndpointsByGroup []portaineree.Endpoint

func (e EndpointsByGroup) Len() int {
	return len(e)
}

func (e EndpointsByGroup) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e EndpointsByGroup) Less(i, j int) bool {
	if e[i].GroupID == e[j].GroupID {
		return false
	}

	groupA := endpointGroupNames[e[i].GroupID]
	groupB := endpointGroupNames[e[j].GroupID]

	return natsort.Compare(strings.ToLower(groupA), strings.ToLower(groupB))
}
