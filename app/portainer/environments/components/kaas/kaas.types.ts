import { Option } from '@/portainer/components/form-components/Input/Select';
import { KaasProvider } from '@/portainer/settings/cloud/types';
import { EnvironmentMetadata } from '@/portainer/environments/environment.service/create';

export interface CloudApiKeys {
  CivoApiKey: string;
  DigitalOceanToken: string;
  LinodeToken: string;
}

export interface KaasProvisionAPIPayload {
  Name: string;
  NodeSize: string;
  NodeCount: number;
  KubernetesVersion: string;
  Region: string;
  NetworkID?: string;
  CredentialID?: number;
  Meta: EnvironmentMetadata;
}

export interface KaasCreateFormValues {
  type: KaasProvider;
  name: string;
  kubernetesVersion: string;
  nodeSize: string;
  nodeCount: number;
  region: string;
  networkId?: string;
  credentialId?: number;
  meta: EnvironmentMetadata;
}

export const KaasCreateFormInitialValues: Pick<
  KaasCreateFormValues,
  'type' | 'name' | 'nodeCount'
> = {
  type: KaasProvider.CIVO,
  name: '',
  nodeCount: 3,
};

type CivoNetwork = {
  Region: string;
  Networks: Array<{ Id: string; Name: string }>;
};

type KaasNodeSize = {
  name: string;
  value: string;
};

type KaasRegion = {
  name: string;
  value: string;
};

export interface KaasInfoResponse {
  Networks?: Array<CivoNetwork>;
  KubernetesVersions: string[];
  NodeSizes: Array<KaasNodeSize>;
  Regions: Array<KaasRegion>;
}

export type NetworkInfo = Map<string, Option<string>[]>;

export interface KaasInfo {
  networks?: NetworkInfo;
  kubernetesVersions: string[];
  nodeSizes: Array<Option<string>>;
  regions: Array<Option<string>>;
}

export type CredentialProviderInfo = Map<string, Array<Option<number>>>;
