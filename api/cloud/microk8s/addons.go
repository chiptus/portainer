package microk8s

import (
	"os"
	"slices"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/database/models"
	pslices "github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/rs/zerolog/log"
)

type AddonsWithArgs []portaineree.MicroK8sAddon

func (a Addons) GetAddonsWithArgs() AddonsWithArgs {
	var addons []portaineree.MicroK8sAddon
	for _, addon := range a {
		if addon.IsAvailable && !addon.IsDefault {
			addons = append(addons, portaineree.MicroK8sAddon{Name: addon.Name})
		}
	}
	return addons
}

func (a Addons) GetNames() []string {
	return pslices.Map(a, func(addon Addon) string { return addon.Name })
}

func (a Addons) GetAddon(name string) *Addon {
	for _, addon := range a {
		if addon.Name == name {
			return &addon
		}
	}

	return nil
}

func (a Addons) IndexOf(element string) int {
	return slices.IndexFunc(a, func(addon Addon) bool { return element == addon.Name })
}

func (addons AddonsWithArgs) GetNames() []string {
	return pslices.Map(addons, func(addon portaineree.MicroK8sAddon) string { return addon.Name })
}

func (addons AddonsWithArgs) EnableAddons(
	masterNodes []string,
	workerNodes []string,
	credential *models.CloudCredential,
	setMessage func(a, b string, c portaineree.EndpointOperationStatus) error,
) AddonsWithArgs {
	log.Info().Msgf("Enabling addons")

	failedAddons := AddonsWithArgs{}

	if len(addons) > 0 {
		allAvailableAddons := GetAllAvailableAddons()

		for _, addon := range addons {
			addonConfig := allAvailableAddons.GetAddon(addon.Name)
			if addonConfig == nil {
				log.Warn().Msgf("addon does not exists in the list of available addons: %s", addon)
				continue
			}

			var ips []string
			switch addonConfig.RequiredOn {
			case "masters":
				ips = masterNodes
			case "all":
				ips = masterNodes
				ips = append(ips, workerNodes...)
			default:
				if len(masterNodes) > 0 {
					ips = append(ips, masterNodes[0])
				}
			}

			log.Debug().Msgf("Enabling addon (%s) on nodes (ips: %s)", addon.Name, strings.Join(ips[:], ", "))
			for _, ip := range ips {
				func() {
					sshClientNode, err := sshUtil.NewConnectionWithCredentials(ip, credential)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to create ssh connection for node %s", ip)
						failedAddons = append(failedAddons, addon)
						return
					}
					defer sshClientNode.Close()

					err = EnableMicrok8sAddonsOnNode(sshClientNode, addon)
					if err != nil {
						// Rather than fail the whole thing. Warn the user and allow them to manually try to enable the addon
						log.Warn().AnErr("error", err).Msgf("failed to enable microk8s addon %s on node. error: ", addon)
						failedAddons = append(failedAddons, addon)
					}
				}()
			}
		}

		if len(failedAddons) > 0 {
			log.Error().Msgf("failed to disable [%v] microk8s addons on node. Please disable these manually", failedAddons)
		}
	}

	return failedAddons
}

func (addons AddonsWithArgs) DisableAddons(
	masterNodes []string,
	workerNodes []string,
	credential *models.CloudCredential,
	setMessage func(a, b string, c portaineree.EndpointOperationStatus) error,
) AddonsWithArgs {

	failedAddons := []portaineree.MicroK8sAddon{}
	log.Info().Msgf("Disabling addons")

	if len(addons) > 0 {
		allAvailableAddons := GetAllAvailableAddons()

		for _, addon := range addons {
			addonConfig := allAvailableAddons.GetAddon(addon.Name)
			if addonConfig == nil {
				log.Warn().Msgf("addon does not exists in the list of available addons: %s", addon)
				failedAddons = append(failedAddons, addon)
				continue
			}

			var ips []string
			switch addonConfig.RequiredOn {
			case "masters":
				ips = masterNodes
			case "all":
				ips = masterNodes
				ips = append(ips, workerNodes...)
			default:
				if len(masterNodes) > 0 {
					ips = append(ips, masterNodes[0])
				}
			}

			log.Debug().Msgf("Disabling addon (%s) on nodes (ips: %s)", addon.Name, strings.Join(ips[:], ", "))
			for _, ip := range ips {
				func() {
					sshClientNode, err := sshUtil.NewConnectionWithCredentials(ip, credential)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to create ssh connection for node %s", ip)
						failedAddons = append(failedAddons, addon)
						return
					}
					defer sshClientNode.Close()

					err = DisableMicrok8sAddonsOnNode(sshClientNode, addon.Name)
					if err != nil {
						// Rather than fail the whole thing. Warn the user and allow them to manually try to disable the addon
						log.Warn().AnErr("error", err).Msgf("failed to disable microk8s addon %s on node. error: ", addon)
						failedAddons = append(failedAddons, addon)
					}

					if addon.Name == "metrics-server" {
						// Wait until the metrics-server pod in the kube-system namespace exits or timeout after 5 minutes
						sshClientNode.RunCommand("microk8s kubectl wait pod -l k8s-app=metrics-server --for=delete -n kube-system --timeout=300s", os.Stdout)
					}
				}()
			}
		}

		if len(failedAddons) > 0 {
			log.Error().Msgf("failed to disable [%v] microk8s addons on node. Please disable these manually", failedAddons)
		}
	}

	return failedAddons
}
