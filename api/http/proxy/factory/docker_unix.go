//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package factory

import (
	"net"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/docker"
)

func (factory ProxyFactory) newOSBasedLocalProxy(path string, endpoint *portaineree.Endpoint) (http.Handler, error) {
	transportParameters := &docker.TransportParameters{
		Endpoint:             endpoint,
		DataStore:            factory.dataStore,
		ReverseTunnelService: factory.reverseTunnelService,
		SignatureService:     factory.signatureService,
		DockerClientFactory:  factory.dockerClientFactory,
		UserActivityService:  factory.userActivityService,
	}

	proxy := &dockerLocalProxy{}

	dockerTransport, err := docker.NewTransport(transportParameters, newSocketTransport(path), factory.gitService)
	if err != nil {
		return nil, err
	}

	proxy.transport = dockerTransport
	return proxy, nil
}

func newSocketTransport(socketPath string) *http.Transport {
	return &http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			return net.Dial("unix", socketPath)
		},
	}
}
