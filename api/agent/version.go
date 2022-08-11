package agent

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	netUrl "net/url"
	"strconv"
	"strings"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
)

// GetAgentVersionAndPlatform returns the agent version and platform
//
// it sends a ping to the agent and parses the version and platform from the headers
func GetAgentVersionAndPlatform(url string, tlsConfig *tls.Config) (portaineree.AgentPlatform, string, error) {
	httpCli := &http.Client{
		Timeout: 3 * time.Second,
	}

	if tlsConfig != nil {
		httpCli.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	if !strings.Contains(url, "//") {
		url = "//" + url
	}

	parsedURL, err := netUrl.Parse(fmt.Sprintf("%s/ping", url))
	if err != nil {
		return 0, "", err
	}

	parsedURL.Scheme = "https"

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return 0, "", err
	}

	resp, err := httpCli.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return 0, "", fmt.Errorf("Failed request with status %d", resp.StatusCode)
	}

	version := resp.Header.Get(portaineree.PortainerAgentHeader)
	if version == "" {
		return 0, "", errors.New("Version Header is missing")
	}

	agentPlatformHeader := resp.Header.Get(portaineree.HTTPResponseAgentPlatform)
	if agentPlatformHeader == "" {
		return 0, "", errors.New("Agent Platform Header is missing")
	}

	agentPlatformNumber, err := strconv.Atoi(agentPlatformHeader)
	if err != nil {
		return 0, "", err
	}

	if agentPlatformNumber == 0 {
		return 0, "", errors.New("Agent platform is invalid")
	}

	return portaineree.AgentPlatform(agentPlatformNumber), version, nil
}
