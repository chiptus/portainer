import { EnvironmentMetadata } from '@/react/portainer/environments/environment.service/create';

export enum K8sDistributionType {
  MICROK8S = 'microk8s',
}

export enum KaasProvider {
  CIVO = 'civo',
  LINODE = 'linode',
  DIGITAL_OCEAN = 'digitalocean',
  GOOGLE_CLOUD = 'gke',
  AWS = 'amazon',
  AZURE = 'azure',
}

export type ProvisionOption = KaasProvider | K8sDistributionType;

export type Microk8sK8sVersion =
  | 'latest/stable'
  | '1.27/stable'
  | '1.26/stable'
  | '1.25/stable'
  | '1.24/stable';

export interface CreateMicrok8sClusterFormValues {
  nodeIPs: string[];
  addons: string[];
  kubernetesVersion: Microk8sK8sVersion;
}

export interface K8sInstallFormValues {
  credentialId: number;
  name: string;
  meta: EnvironmentMetadata;

  microk8s: CreateMicrok8sClusterFormValues;
}

export interface CreateMicrok8sClusterPayload
  extends CreateMicrok8sClusterFormValues {
  name: string;
}

export const providerTitles: Record<KaasProvider, string> = {
  civo: 'Civo',
  linode: 'Linode',
  digitalocean: 'DigitalOcean',
  gke: 'Google Cloud',
  amazon: 'AWS',
  azure: 'Azure',
};

export const k8sInstallTitles: Record<K8sDistributionType, string> = {
  microk8s: 'MicroK8s',
};

export const provisionOptionTitles: Record<ProvisionOption, string> = {
  ...providerTitles,
  ...k8sInstallTitles,
};
