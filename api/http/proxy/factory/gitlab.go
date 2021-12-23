package factory

import (
	"net/http"
	"net/url"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/gitlab"
)

func newGitlabProxy(uri string, userActivityService portaineree.UserActivityService) (http.Handler, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	proxy := newSingleHostReverseProxyWithHostHeader(url)
	proxy.Transport = gitlab.NewTransport(userActivityService)
	return proxy, nil
}
