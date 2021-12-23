package testhelpers

import portaineree "github.com/portainer/portainer-ee/api"

type ReverseTunnelService struct{}

func (r ReverseTunnelService) StartTunnelServer(addr, port string, snapshotService portaineree.SnapshotService) error {
	return nil
}
func (r ReverseTunnelService) GenerateEdgeKey(url, host string, endpointIdentifier int) string {
	return "nil"
}
func (r ReverseTunnelService) SetTunnelStatusToActive(endpointID portaineree.EndpointID) {}
func (r ReverseTunnelService) SetTunnelStatusToRequired(endpointID portaineree.EndpointID) error {
	return nil
}
func (r ReverseTunnelService) SetTunnelStatusToIdle(endpointID portaineree.EndpointID) {}
func (r ReverseTunnelService) GetTunnelDetails(endpointID portaineree.EndpointID) *portaineree.TunnelDetails {
	return nil
}
func (r ReverseTunnelService) AddEdgeJob(endpointID portaineree.EndpointID, edgeJob *portaineree.EdgeJob) {
}
func (r ReverseTunnelService) RemoveEdgeJob(edgeJobID portaineree.EdgeJobID) {}
