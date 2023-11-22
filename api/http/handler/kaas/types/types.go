package types

import (
	"fmt"
	"net/http"
	"slices"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

type (
	CloudProviderShortName string

	CloudProviders map[CloudProviderShortName]portaineree.CloudProvider

	EnvironmentMetadata struct {
		GroupId               portainer.EndpointGroupID  `json:"groupId"`
		TagIds                []portainer.TagID          `json:"tagIds"`
		CustomTemplateID      portainer.CustomTemplateID `json:"customTemplateID"`
		CustomTemplateContent string                     `json:"customTemplateContent"`
	}

	Microk8sAddonsPayload struct {
		Addons []portaineree.MicroK8sAddon `json:"addons"`
	}
)

var CloudProvidersMap CloudProviders = CloudProviders{
	portaineree.CloudProviderCivo: {
		Provider: portaineree.CloudProviderCivo,
		Name:     "Civo",
		URL:      "https://www.civo.com/login",
	},
	portaineree.CloudProviderLinode: {
		Provider: portaineree.CloudProviderLinode,
		Name:     "Linode",
		URL:      "https://login.linode.com/",
	},
	portaineree.CloudProviderDigitalOcean: {
		Provider: portaineree.CloudProviderDigitalOcean,
		Name:     "Digital Ocean",
		URL:      "https://cloud.digitalocean.com/login",
	},
	portaineree.CloudProviderGKE: {
		Provider: portaineree.CloudProviderGKE,
		Name:     "Google Cloud Platform",
		URL:      "https://console.cloud.google.com/kubernetes/",
	},
	portaineree.CloudProviderAzure: {
		Provider: portaineree.CloudProviderAzure,
		Name:     "Azure",
		URL:      "https://portal.azure.com/",
	},
	portaineree.CloudProviderAmazon: {
		Provider: portaineree.CloudProviderAmazon,
		Name:     "Amazon",
		URL:      "https://console.aws.amazon.com",
	},
	portaineree.CloudProviderMicrok8s: {
		Provider: portaineree.CloudProviderMicrok8s,
		Name:     "MicroK8s",
		URL:      "https://microk8s.io/",
	},
	portaineree.CloudProviderKubeConfig: {
		Provider: portaineree.CloudProviderKubeConfig,
		Name:     "KubeConfig",
		URL:      "",
	},
}

func (r Microk8sAddonsPayload) Validate(request *http.Request) error {
	if len(r.Addons) == 0 {
		return fmt.Errorf("at least one addon must be specified")
	}

	// TODO: this creates an import loop.  Refactor to avoid.
	// for _, addon := range r.Addons {
	// 	if !slices.Contains(microk8s.Microk8sDefaultInstallableAddons, addon) {
	// 		return fmt.Errorf("the specified addon (%s) is not valid or cannot be installed by portainer", addon)
	// 	}
	// }

	return nil
}

func (r Microk8sAddonsPayload) IndexOf(element string) int {
	return slices.IndexFunc(r.Addons, func(addon portaineree.MicroK8sAddon) bool { return element == addon.Name })
}
