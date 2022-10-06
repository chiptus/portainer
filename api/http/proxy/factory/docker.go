package factory

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/docker"
	"github.com/portainer/portainer-ee/api/internal/url"
	"github.com/portainer/portainer/api/crypto"

	"github.com/rs/zerolog/log"
)

func (factory *ProxyFactory) newDockerProxy(endpoint *portaineree.Endpoint) (http.Handler, error) {
	if strings.HasPrefix(endpoint.URL, "unix://") || strings.HasPrefix(endpoint.URL, "npipe://") {
		return factory.newDockerLocalProxy(endpoint)
	}

	return factory.newDockerHTTPProxy(endpoint)
}

func (factory *ProxyFactory) newDockerLocalProxy(endpoint *portaineree.Endpoint) (http.Handler, error) {
	endpointURL, err := url.ParseURL(endpoint.URL)
	if err != nil {
		return nil, err
	}

	return factory.newOSBasedLocalProxy(endpointURL.Path, endpoint)
}

func (factory *ProxyFactory) newDockerHTTPProxy(endpoint *portaineree.Endpoint) (http.Handler, error) {
	rawURL := endpoint.URL
	if endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment {
		tunnel := factory.reverseTunnelService.GetTunnelDetails(endpoint.ID)
		rawURL = fmt.Sprintf("http://127.0.0.1:%d", tunnel.Port)
	}

	endpointURL, err := url.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}

	endpointURL.Scheme = "http"
	httpTransport := &http.Transport{}

	if endpoint.TLSConfig.TLS || endpoint.TLSConfig.TLSSkipVerify {
		config, err := crypto.CreateTLSConfigurationFromDisk(endpoint.TLSConfig.TLSCACertPath, endpoint.TLSConfig.TLSCertPath, endpoint.TLSConfig.TLSKeyPath, endpoint.TLSConfig.TLSSkipVerify)
		if err != nil {
			return nil, err
		}

		httpTransport.TLSClientConfig = config
		endpointURL.Scheme = "https"
	}

	transportParameters := &docker.TransportParameters{
		Endpoint:             endpoint,
		DataStore:            factory.dataStore,
		ReverseTunnelService: factory.reverseTunnelService,
		SignatureService:     factory.signatureService,
		DockerClientFactory:  factory.dockerClientFactory,
		UserActivityService:  factory.userActivityService,
	}

	dockerTransport, err := docker.NewTransport(transportParameters, httpTransport, factory.gitService)
	if err != nil {
		return nil, err
	}

	proxy := newSingleHostReverseProxyWithHostHeader(endpointURL)
	proxy.Transport = dockerTransport
	return proxy, nil
}

type dockerLocalProxy struct {
	transport *docker.Transport
}

// ServeHTTP is the http.Handler interface implementation
// for a local (Unix socket or Windows named pipe) Docker proxy.
func (proxy *dockerLocalProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Force URL/domain to http/unixsocket to be able to
	// use http.transport RoundTrip to do the requests via the socket
	r.URL.Scheme = "http"
	r.URL.Host = "unixsocket"

	res, err := proxy.transport.ProxyDockerRequest(r)
	if err != nil {
		code := http.StatusInternalServerError
		if res != nil && res.StatusCode != 0 {
			code = res.StatusCode
		}

		httperror.WriteError(w, code, "Unable to proxy the request via the Docker socket", err)
		return
	}
	defer res.Body.Close()

	for k, vv := range res.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(res.StatusCode)

	if _, err := io.Copy(w, res.Body); err != nil {
		log.Debug().Err(err).Msg("proxy error")
	}
}
