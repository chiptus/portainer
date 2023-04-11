package podsecurity

import (
	"context"
	"path"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
)

type GateKeeper struct {
	kubernetesDeployer portaineree.KubernetesDeployer
	assetsPath         string
}

func NewGateKeeper(kubernetesDeployer portaineree.KubernetesDeployer, assetsPath string) *GateKeeper {
	return &GateKeeper{
		kubernetesDeployer: kubernetesDeployer,
		assetsPath:         assetsPath,
	}
}

func (g *GateKeeper) Deploy(userID portaineree.UserID, endpoint *portaineree.Endpoint) error {
	manifest := path.Join(g.assetsPath, "pod-security-policy", GateKeeperFile)
	_, err := g.kubernetesDeployer.Deploy(userID, endpoint, []string{manifest}, GateKeeperNameSpace)
	if err != nil {
		log.Debug().Err(err).Msg("failed to deploy kubernetes gatekeeper")
		log.Error().Msg("failed to deploy kubernetes gatekeeper, remove installed files")
		g.Remove(userID, endpoint)
		return err
	}
	return nil
}

func (g *GateKeeper) Remove(userID portaineree.UserID, endpoint *portaineree.Endpoint) error {
	manifest := path.Join(g.assetsPath, "pod-security-policy", GateKeeperFile)
	_, err := g.kubernetesDeployer.Remove(userID, endpoint, []string{manifest}, GateKeeperNameSpace)
	if err != nil {
		log.Debug().Err(err).Msg("failed to remove kubernetes gatekeeper")
		log.Error().Msg("failed to remove kubernetes gatekeeper")
		return err
	}
	return nil
}

func (g *GateKeeper) DeployExcludedNamespaces(userID portaineree.UserID, endpoint *portaineree.Endpoint) error {
	gatekeeperExcludedNamespacesManifest := path.Join(g.assetsPath, "pod-security-policy", GateKeeperExcludedNamespacesFile)
	_, err := g.kubernetesDeployer.Deploy(userID, endpoint, []string{gatekeeperExcludedNamespacesManifest}, GateKeeperNameSpace)
	if err != nil {
		log.Debug().Err(err).Msg("failed to deploy kubernetes gatekeeper excluded namespaces")
		log.Error().Msg("failed to deploy kubernetes gatekeeper excluded namespaces, remove installed files")
		g.Remove(userID, endpoint)
		return err
	}
	return nil
}

func (g *GateKeeper) RemoveExcludedNamespaces(userID portaineree.UserID, endpoint *portaineree.Endpoint) error {
	gatekeeperExcludedNamespacesManifest := path.Join(g.assetsPath, "pod-security-policy", GateKeeperExcludedNamespacesFile)
	_, err := g.kubernetesDeployer.Remove(userID, endpoint, []string{gatekeeperExcludedNamespacesManifest}, GateKeeperNameSpace)
	if err != nil {
		log.Debug().Err(err).Msg("failed to remove kubernetes gatekeeper excluded namespaces")
		log.Error().Msg("failed to remove kubernetes gatekeeper excluded namespaces")
		return err
	}
	return nil
}

func (g *GateKeeper) DeployPodSecurityConstraints(userID portaineree.UserID, endpoint *portaineree.Endpoint) error {
	for _, v := range PodSecurityConstraintsMap {
		log.Info().Str("template", v).Msg("deploying constraint template")

		_, err := g.kubernetesDeployer.Deploy(userID, endpoint, []string{path.Join(g.assetsPath, "pod-security-policy", v, "template.yaml")}, GateKeeperNameSpace)
		if err != nil {
			log.Error().Str("template", v).Err(err).Msg("unable to deploy")

			return err
		}

		log.Info().Str("template", v).Msg("successfully deployed")
	}

	return nil
}

func (g *GateKeeper) UpgradeEndpoint(userID portaineree.UserID, endpoint *portaineree.Endpoint, kubeclient portaineree.KubeClient, kubeClientSet *kubernetes.Clientset, exitingRule *PodSecurityRule) error {

	_, err := kubeclient.GetNamespaces()
	if err != nil {
		log.Error().Msgf("Updating gatekeeper. error connecting endpoint (%d): %s", endpoint.ID, err)
		return err
	}

	log.Info().Msgf("Updating gatekeeper for endpoint: %d", endpoint.ID)
	// if the rule is enabled, we need to update the rule to the latest version
	//1.deploy gatekeeper
	err = g.Deploy(userID, endpoint)
	if err != nil {
		return err
	}

	err = WaitForOpaReady(context.Background(), kubeClientSet)
	if err != nil {
		return err
	}

	//2. deploy gatekeeper excluded namespaces
	err = g.DeployExcludedNamespaces(userID, endpoint)
	if err != nil {
		return err
	}

	//3.deploy gatekeeper constrainttemplate
	err = g.DeployPodSecurityConstraints(userID, endpoint)
	if err != nil {
		return err
	}

	//3.deploy gatekeeper constraint yaml files
	for name := range PodSecurityConstraintsMap {
		rulename := name
		constraint := PodSecurityConstraint{}
		constraint.Init(userID, endpoint, rulename, exitingRule, exitingRule, g.assetsPath)

		if err := constraint.Fresh(g.kubernetesDeployer); err != nil {
			log.Error().Str("rule", rulename).Err(err).Msg("unable to deploy constraint")
		}

		log.Debug().Str("rule", rulename).Msg("successfully deployed constraint")
	}
	log.Info().Msgf("End updating gatekeeper for endpoint: %d", endpoint.ID)

	return nil
}
