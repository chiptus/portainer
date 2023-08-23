import {
  AddOnFormValue,
  AddonsArgumentType,
} from '@/react/kubernetes/cluster/microk8s/addons/types';
import { EnvironmentMetadata } from '@/react/portainer/environments/environment.service/create';

import { Option } from '@@/form-components/Input/Select';

import { K8sDistributionType, KaasProvider } from '../../../types';

export type ProvisionOption = KaasProvider | K8sDistributionType;

export type AddonOptionInfo = {
  versionAvailableFrom: string;
  versionAvailableTo: string;
  repository: string;

  tooltip?: string;
  info?: string;
  version?: string;
  placeholder?: string;
  argumentsType: AddonsArgumentType;
  isDefault: boolean;
} & Option<string>;

export interface MicroK8sInfo {
  kubernetesVersions: Option<string>[];
  availableAddons: AddonOptionInfo[];
  requiredAddons: string[];
}

export interface CreateMicrok8sClusterFormValues {
  masterNodes: string[];
  workerNodes: string[];
  addons: AddOnFormValue[];
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
  addons: AddOnFormValue[];
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
