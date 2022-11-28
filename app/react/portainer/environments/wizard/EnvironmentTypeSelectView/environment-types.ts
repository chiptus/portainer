import { FeatureId } from '@/react/portainer/feature-flags/enums';
import DockerIcon from '@/assets/ico/vendor/docker-icon.svg?c';
import Kube from '@/assets/ico/kube.svg?c';
import MicrosoftIcon from '@/assets/ico/vendor/microsoft-icon.svg?c';
import NomadIcon from '@/assets/ico/vendor/nomad-icon.svg?c';

import KaaSIcon from './kaas-icon.svg?c';

export const environmentTypes = [
  {
    id: 'dockerStandalone',
    title: 'Docker Standalone',
    formTitle: 'Connect to your Docker Standalone environment',
    icon: DockerIcon,
    description: 'Connect to Docker Standalone via URL/IP, API or Socket',
  },
  {
    id: 'dockerSwarm',
    title: 'Docker Swarm',
    formTitle: 'Connect to your Docker Swarm environment',
    icon: DockerIcon,
    description: 'Connect to Docker Swarm via URL/IP, API or Socket',
  },
  {
    id: 'kubernetes',
    title: 'Kubernetes',
    formTitle: 'Connect to your Kubernetes environment',
    icon: Kube,
    description: 'Connect to a kubernetes environment via URL/IP',
  },
  {
    id: 'aci',
    title: 'ACI',
    formTitle: 'Connect to your ACI environment',
    description: 'Connect to ACI environment via API',
    icon: MicrosoftIcon,
  },
  {
    id: 'nomad',
    title: 'Nomad',
    formTitle: 'Connect to your Nomad environment',
    description: 'Connect to HashiCorp Nomad environment via API',
    icon: NomadIcon,
    featureId: FeatureId.NOMAD,
  },
  {
    id: 'kaas',
    title: 'KaaS',
    formTitle: 'Provision a KaaS environment',
    description: 'Provision a Kubernetes environment with a cloud provider',
    icon: KaaSIcon,
    featureId: FeatureId.KAAS_PROVISIONING,
  },
] as const;
