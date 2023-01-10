package chisel

import (
	"encoding/base64"
	"strconv"
	"strings"
)

// GenerateEdgeKey will generate a key that can be used by an Edge agent to register with a Portainer instance.
// The key represents the following data in this particular format:
// portainer_api_server_url|tunnel_server_addr|tunnel_server_fingerprint|endpoint_ID
// The key returned by this function is a base64 encoded version of the data.
func (service *Service) GenerateEdgeKey(apiURL, tunnelAddr string, endpointIdentifier int) string {
	keyInformation := []string{
		apiURL,
		tunnelAddr,
		service.serverFingerprint,
		strconv.Itoa(endpointIdentifier),
	}

	key := strings.Join(keyInformation, "|")
	return base64.RawStdEncoding.EncodeToString([]byte(key))
}
