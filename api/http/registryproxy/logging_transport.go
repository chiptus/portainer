package registryproxy

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/useractivity"
	ru "github.com/portainer/portainer-ee/api/http/utils"
)

type loggingTransport struct {
	transport           http.RoundTripper
	userActivityService portaineree.UserActivityService
}

func NewLoggingTransport(userActivityService portaineree.UserActivityService, transport http.RoundTripper) *loggingTransport {
	return &loggingTransport{
		transport:           transport,
		userActivityService: userActivityService,
	}
}

// RoundTrip satisfies http.RoundTripper interface
// it proxies the request to the underlying roundtripper and logs the request
func (lt *loggingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// need a copy of the request body to preserve the original
	body := ru.CopyRequestBody(request)

	response, err := lt.transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	useractivity.LogProxiedActivity(lt.userActivityService, nil, response.StatusCode, body, request)

	return response, err
}
