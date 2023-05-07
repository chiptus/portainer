package types

import portaineree "github.com/portainer/portainer-ee/api"

type (
	CloudProviderShortName string

	CloudProviders map[CloudProviderShortName]portaineree.CloudProvider

	EnvironmentMetadata struct {
		GroupId               portaineree.EndpointGroupID  `json:"groupId"`
		TagIds                []portaineree.TagID          `json:"tagIds"`
		CustomTemplateID      portaineree.CustomTemplateID `json:"customTemplateID"`
		CustomTemplateContent string                       `json:"customTemplateContent"`
	}
)

var CloudProvidersMap CloudProviders = CloudProviders{
	portaineree.CloudProviderCivo: {
		Name: "Civo",
		URL:  "https://www.civo.com/login",
	},
	portaineree.CloudProviderLinode: {
		Name: "Linode",
		URL:  "https://login.linode.com/",
	},
	portaineree.CloudProviderDigitalOcean: {
		Name: "DigitalOcean",
		URL:  "https://cloud.digitalocean.com/login",
	},
	portaineree.CloudProviderGKE: {
		Name: "Google Cloud Platform",
		URL:  "https://console.cloud.google.com/kubernetes/",
	},
	portaineree.CloudProviderAzure: {
		Name: "Azure",
		URL:  "https://portal.azure.com/",
	},
	portaineree.CloudProviderAmazon: {
		Name: "Amazon",
		URL:  "https://console.aws.amazon.com",
	},
	portaineree.CloudProviderMicrok8s: {
		Name: "MicroK8s",
		URL:  "https://microk8s.io/",
	},
}
