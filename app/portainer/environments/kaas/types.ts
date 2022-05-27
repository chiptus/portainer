import { Option } from '@/portainer/components/form-components/Input/Select';
import { EnvironmentMetadata } from '@/portainer/environments/environment.service/create';

export interface CloudApiKeys {
  CivoApiKey: string;
  DigitalOceanToken: string;
  LinodeToken: string;
}

// Form values
interface CreateBaseClusterFormValues {
  kubernetesVersion: string;
  region: string;
  credentialId: number;
  meta: EnvironmentMetadata;
  nodeCount: number;
}

export interface CreateApiClusterFormValues
  extends CreateBaseClusterFormValues {
  nodeSize: string;
  networkId?: string;
}

export interface CreateGKEClusterFormValues
  extends CreateBaseClusterFormValues {
  nodeSize: string;
  cpu: number;
  ram: number;
  hdd: number;
  networkId: string;
}

// Create KaaS cluster payloads
export interface CreateApiClusterPayload extends CreateApiClusterFormValues {
  name: string;
}

export interface CreateGKEClusterPayload extends CreateBaseClusterFormValues {
  name: string;
  nodeSize: string;
  cpu?: number;
  ram?: number;
  hdd: number;
  networkId: string;
}

// Kaas info response
type KaasNetwork = {
  region: string;
  networks: Array<{ id: string; name: string }>;
};

type KaasNodeSize = {
  name: string;
  value: string;
};

type KaasRegion = {
  name: string;
  value: string;
};

type KubernetesVersion = {
  name: string;
  value: string;
};

interface CPUInfo {
  // the number of cpu cores per node
  default: number;
  min: number;
  max: number;
}

interface RAMInfo {
  // the amount of RAM per node
  default: number;
  min: number;
  max: number;
}

interface HDDInfo {
  // the amount of disk space per node
  default: number;
  min: number;
  max: number;
}

interface KaasBaseInfoResponse {
  regions: Array<KaasRegion>;
  kubernetesVersions: KubernetesVersion[];
}

export interface KaasApiInfoResponse extends KaasBaseInfoResponse {
  networks?: Array<KaasNetwork>;
  nodeSizes: Array<KaasNodeSize>;
}

export interface KaasGKEInfoResponse extends KaasBaseInfoResponse {
  cpu: CPUInfo;
  hdd: HDDInfo;
  ram: RAMInfo;
  networks: Array<KaasNetwork>;
  nodeSizes: Array<KaasNodeSize>;
}

export type KaasInfoResponse = KaasApiInfoResponse | KaasGKEInfoResponse;

// returns true if the response is a api info response
export function isAPIKaasInfoResponse(
  kaasInfoResponse: KaasInfoResponse
): kaasInfoResponse is KaasApiInfoResponse {
  return 'nodeSizes' in kaasInfoResponse;
}

// returns true if the response is a gke info response
export function isGKEKaasInfoResponse(
  kaasInfoResponse: KaasInfoResponse
): kaasInfoResponse is KaasGKEInfoResponse {
  return 'cpu' in kaasInfoResponse;
}

// Formatted Kaas info
type NetworkInfo = Map<string, Option<string>[]>;

interface BaseKaasInfo {
  kubernetesVersions: Array<Option<string>>;
  regions: Array<Option<string>>;
}

export interface APIKaasInfo extends BaseKaasInfo {
  networks?: NetworkInfo;
  nodeSizes: Array<Option<string>>;
}

export interface GKEKaasInfo extends BaseKaasInfo {
  nodeSizes: Array<Option<string>>;
  networks: NetworkInfo;
  cpu: CPUInfo;
  hdd: HDDInfo;
  ram: RAMInfo;
}

export type KaasInfo = APIKaasInfo | GKEKaasInfo;

export function isAPIKaasInfo(kaasInfo: KaasInfo): kaasInfo is APIKaasInfo {
  return 'nodeSizes' in kaasInfo;
}

export function isGKEKaasInfo(kaasInfo: KaasInfo): kaasInfo is GKEKaasInfo {
  return 'cpu' in kaasInfo;
}

export type CredentialProviderInfo = Map<string, Array<Option<number>>>;
