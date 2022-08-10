import { FeatureId } from '@/portainer/feature-flags/enums';

import KaaSIcon from './kaas-icon.svg?c';

export const environmentTypes = [
  {
    id: 'docker',
    title: 'Docker',
    formTitle: 'Connect to your Docker environment',
    icon: 'fab fa-docker',
    description:
      'Connect to Docker Standalone / Swarm via URL/IP, API or Socket',
  },
  {
    id: 'kubernetes',
    title: 'Kubernetes',
    formTitle: 'Connect to your Kubernetes environment',
    icon: 'fas fa-dharmachakra',
    description: 'Connect to a kubernetes environment via URL/IP',
  },
  {
    id: 'aci',
    title: 'ACI',
    formTitle: 'Connect to your ACI environment',
    description: 'Connect to ACI environment via API',
    icon: 'fab fa-microsoft',
  },
  {
    id: 'nomad',
    title: 'Nomad',
    formTitle: 'Connect to your Nomad environment',
    description: 'Connect to HashiCorp Nomad environment via API',
    icon: 'nomad-icon',
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
