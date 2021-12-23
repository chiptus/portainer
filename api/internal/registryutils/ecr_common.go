package registryutils

import (
	"fmt"
	"regexp"

	portaineree "github.com/portainer/portainer-ee/api"
)

var registryIdRule = regexp.MustCompile(`^(?P<registryid>[0-9]{12}).*`)

func GetRegistryId(registry *portaineree.Registry) (registryId string, err error) {
	if registry.Type != portaineree.EcrRegistry {
		err = fmt.Errorf("invalid registry type to get ECR registry ID")
		return
	}

	if registryIdRule.MatchString(registry.URL) {
		match := registryIdRule.FindStringSubmatch(registry.URL)
		registryId = match[1]
	} else {
		err = fmt.Errorf("invalid registry URL to get ECR registry ID")
	}

	return
}
