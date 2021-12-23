package websocket

import portaineree "github.com/portainer/portainer-ee/api"

type webSocketRequestParams struct {
	ID       string
	nodeName string
	endpoint *portaineree.Endpoint
	token    string
}
