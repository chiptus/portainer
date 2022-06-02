import { Option } from '@/portainer/components/form-components/Input/Select';
import { EnvironmentMetadata } from '@/portainer/environments/environment.service/create';

// Form values
interface CreateBaseClusterFormValues {
  kubernetesVersion: string;
  region: string;
  credentialId: number;
  meta: EnvironmentMetadata;
  nodeCount: number;
  nodeSize: string;
}

export interface CreateApiClusterFormValues
  extends CreateBaseClusterFormValues {
  networkId?: string;
}

export interface CreateAzureClusterFormValues
  extends CreateBaseClusterFormValues {
  resourceGroup: string;
  resourceGroupName?: string;
  tier: string;
  poolName: string;
  dnsPrefix: string;
  availabilityZones: string[];
  resourceGroupInput: string;
}

export interface CreateGKEClusterFormValues
  extends CreateBaseClusterFormValues {
  cpu: number;
  ram: number;
  hdd: number;
  networkId: string;
}

export function isApiClusterFormValues(
  values: CreateBaseClusterFormValues
): values is CreateApiClusterFormValues {
  return !('cpu' in values) && !('availabilityZones' in values);
}

export function isAzureClusterFormValues(
  values: CreateBaseClusterFormValues
): values is CreateAzureClusterFormValues {
  return 'resourceGroup' in values;
}

export function isGkeClusterFormValues(
  values: CreateBaseClusterFormValues
): values is CreateGKEClusterFormValues {
  return 'cpu' in values;
}

export type CreateClusterFormValues =
  | CreateApiClusterFormValues
  | CreateAzureClusterFormValues
  | CreateGKEClusterFormValues;

// Create KaaS cluster payloads
export interface CreateApiClusterPayload extends CreateApiClusterFormValues {
  name: string;
}

export interface CreateAzureClusterPayload
  extends CreateAzureClusterFormValues {
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

export type CreateClusterPayload =
  | CreateApiClusterPayload
  | CreateAzureClusterPayload
  | CreateGKEClusterPayload;

// Kaas info response
type KaasNetwork = {
  region: string;
  networks: { id: string; name: string }[];
};

type NodeSize = {
  name: string;
  value: string;
};

type AzureNodeSize = {
  name: string;
  value: string;
  zones: string[] | null;
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
  regions: KaasRegion[];
  kubernetesVersions: KubernetesVersion[];
}

export interface KaasApiInfoResponse extends KaasBaseInfoResponse {
  networks?: KaasNetwork[];
  nodeSizes: NodeSize[];
}

export interface KaasAzureInfoResponse extends KaasBaseInfoResponse {
  nodeSizes: Record<string, AzureNodeSize[]>; // key is the region
  resourceGroups: string[];
  tiers: string[];
}

export interface KaasGKEInfoResponse extends KaasBaseInfoResponse {
  cpu: CPUInfo;
  hdd: HDDInfo;
  ram: RAMInfo;
  networks: Array<KaasNetwork>;
  nodeSizes: Array<KaasNodeSize>;
}

export type KaasInfoResponse =
  | KaasApiInfoResponse
  | KaasGKEInfoResponse
  | KaasAzureInfoResponse;

// returns true if the response is a api info response
export function isAPIKaasInfoResponse(
  kaasInfoResponse: KaasInfoResponse
): kaasInfoResponse is KaasApiInfoResponse {
  return 'nodeSizes' in kaasInfoResponse;
}

// returns true if the response is a gke info response
export function isAzureKaasInfoResponse(
  kaasInfoResponse: KaasInfoResponse
): kaasInfoResponse is KaasAzureInfoResponse {
  return 'resourceGroups' in kaasInfoResponse;
}

export function isGKEKaasInfoResponse(
  kaasInfoResponse: KaasInfoResponse
): kaasInfoResponse is KaasGKEInfoResponse {
  return 'cpu' in kaasInfoResponse;
}

// Formatted Kaas info
export type NetworkInfo = Record<string, Option<string>[]>;

interface BaseKaasInfo {
  kubernetesVersions: Option<string>[];
  regions: Option<string>[];
}

export interface APIKaasInfo extends BaseKaasInfo {
  networks?: NetworkInfo;
  nodeSizes: Option<string>[];
}

export interface AzureKaasInfo extends BaseKaasInfo {
  nodeSizes: Record<string, AzureNodeSize[]>;
  resourceGroups: Option<string>[];
  tiers: string[];
}

export interface GKEKaasInfo extends BaseKaasInfo {
  nodeSizes: Array<Option<string>>;
  networks: NetworkInfo;
  cpu: CPUInfo;
  hdd: HDDInfo;
  ram: RAMInfo;
}

export type KaasInfo = APIKaasInfo | AzureKaasInfo | GKEKaasInfo;

export function isAPIKaasInfo(kaasInfo: KaasInfo): kaasInfo is APIKaasInfo {
  return 'nodeSizes' in kaasInfo;
}

export function isAzureKaasInfo(kaasInfo: KaasInfo): kaasInfo is AzureKaasInfo {
  return 'resourceGroups' in kaasInfo;
}

export function isGKEKaasInfo(kaasInfo: KaasInfo): kaasInfo is GKEKaasInfo {
  return 'cpu' in kaasInfo;
}

export type CredentialProviderInfo = Map<string, Option<number>[]>;
