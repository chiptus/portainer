package factory

import (
	"net/http"
	"net/url"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/gitlab"
)

func newGitlabProxy(uri string, userActivityService portainer.UserActivityService) (http.Handler, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	proxy := newSingleHostReverseProxyWithHostHeader(url)
	proxy.Transport = gitlab.NewTransport(userActivityService)
	return proxy, nil
}
