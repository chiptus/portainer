package stackutils

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/stretchr/testify/assert"
)

func TestIsValidStackFile_DefaultPortEnvSubstitution(t *testing.T) {
	yamlContent := []byte(`
version: "3"

services:
  webservice:
    image: nginx
    container_name: hello-world
    networks:
      - "mynet1"
    ports:
      - "${PORT:-8080}:80"

networks:
  mynet1:
    driver: bridge
    ipam:
      config:
        - subnet: 172.16.0.0/24
`)

	securitySettings := &portaineree.EndpointSecuritySettings{}
	err := IsValidStackFile(yamlContent, securitySettings)
	assert.NoError(t, err)
}

func TestIsValidStackFile_PortEnv(t *testing.T) {
	yamlContent := []byte(`
version: "3"

services:
  webservice:
    image: nginx
    container_name: hello-world
    networks:
      - "mynet1"
    ports:
      - "${PORT}:80"

networks:
  mynet1:
    driver: bridge
    ipam:
      config:
        - subnet: 172.16.0.0/24
`)

	securitySettings := &portaineree.EndpointSecuritySettings{}
	err := IsValidStackFile(yamlContent, securitySettings)
	assert.NoError(t, err)
}
