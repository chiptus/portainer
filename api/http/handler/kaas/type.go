package kaas

import portaineree "github.com/portainer/portainer-ee/api"

type (
	CloudProviderShortName string

	CloudProviders map[CloudProviderShortName]portaineree.CloudProvider
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
}
