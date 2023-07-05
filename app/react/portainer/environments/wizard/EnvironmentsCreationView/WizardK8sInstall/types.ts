import { EnvironmentMetadata } from '@/react/portainer/environments/environment.service/create';

import { Option } from '@@/form-components/Input/Select';

import { AddOnOption } from './Microk8sCreateClusterForm/AddonSelector';

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

export type AddonOption = {
  versionAvailableFrom: string;
  type: string;
} & Option<string>;

export interface MicroK8sInfo {
  kubernetesVersions: Option<string>[];
  availableAddons: AddonOption[];
  requiredAddons: string[];
}

export interface CreateMicrok8sClusterFormValues {
  masterNodes: string[];
  workerNodes: string[];
  addons: AddOnOption[];
  kubernetesVersion: string;
}

export interface K8sInstallFormValues {
  credentialId: number;
  name: string;
  meta: EnvironmentMetadata;

  microk8s: CreateMicrok8sClusterFormValues;
}

export interface CreateMicrok8sClusterPayload {
  name: string;
  masterNodes: string[];
  workerNodes: string[];
  addons: string[];
  kubernetesVersion: string;
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
