import { Option } from '@/portainer/components/form-components/Input/Select';
import { EnvironmentMetadata } from '@/portainer/environments/environment.service/create';

interface CreateBaseClusterFormValues {
  kubernetesVersion: string;
  region: string;
  credentialId: number;
  nodeCount: number;
  nodeSize: string;
}

export interface CreateApiClusterFormValues {
  networkId: string;
}

export interface CreateAzureClusterFormValues {
  resourceGroup: string;
  resourceGroupName?: string;
  tier: string;
  poolName: string;
  dnsPrefix: string;
  availabilityZones: string[];
  resourceGroupInput: string;
}

export interface CreateGKEClusterFormValues {
  cpu: number;
  ram: number;
  hdd: number;
  networkId: string;
}

export interface CreateEKSClusterFormValues {
  amiType: string;
  instanceType: string;
  nodeVolumeSize: number;
}

export interface FormValues extends CreateBaseClusterFormValues {
  name: string;
  meta: EnvironmentMetadata;

  azure: CreateAzureClusterFormValues;
  google: CreateGKEClusterFormValues;
  api: CreateApiClusterFormValues;
  amazon: CreateEKSClusterFormValues;
}

// Create KaaS cluster payloads
export type CreateApiClusterPayload = CreateApiClusterFormValues;

export interface CreateAzureClusterPayload
  extends CreateAzureClusterFormValues {
  name: string;
}

export interface CreateGKEClusterPayload extends CreateBaseClusterFormValues {
  cpu?: number;
  ram?: number;
  hdd: number;
  networkId: string;
}

export interface CreateEksClusterPayload extends CreateEKSClusterFormValues {
  name: string;
}

export type CreateClusterPayload =
  | CreateApiClusterPayload
  | CreateEksClusterPayload
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

// the number of cpu cores per node
interface CPUInfo {
  default: number;
  min: number;
  max: number;
}

// the amount of RAM per node
interface RAMInfo {
  default: number;
  min: number;
  max: number;
}

// the amount of disk space per node
interface HDDInfo {
  default: number;
  min: number;
  max: number;
}

interface KaasBaseInfoResponse {
  regions: KaasRegion[];
  kubernetesVersions: KubernetesVersion[];
}

type InstanceType = {
  name: string;
  value: string;
  compatibleAmiTypes: string[];
};

type AMIType = {
  name: string;
  value: string;
};

export type InstanceTypeRegions = Record<string, InstanceType[]>;

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

export interface KaasEksInfoResponse extends KaasBaseInfoResponse {
  amiTypes: AMIType[];
  instanceTypes: InstanceTypeRegions;
}

export type KaasInfoResponse =
  | KaasApiInfoResponse
  | KaasGKEInfoResponse
  | KaasAzureInfoResponse
  | KaasEksInfoResponse;

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

export function isEksKaasInfoResponse(
  kaasInfoResponse: KaasInfoResponse
): kaasInfoResponse is KaasEksInfoResponse {
  return 'amiTypes' in kaasInfoResponse;
}

// KaaS cluster info
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

export interface EKSKaasInfo extends BaseKaasInfo {
  amiTypes: Array<Option<string>>;
  instanceTypes: InstanceTypeRegions;
}

export type KaasInfo = APIKaasInfo | AzureKaasInfo | GKEKaasInfo | EKSKaasInfo;

export function isAPIKaasInfo(kaasInfo: KaasInfo): kaasInfo is APIKaasInfo {
  return 'nodeSizes' in kaasInfo;
}

export function isAzureKaasInfo(kaasInfo: KaasInfo): kaasInfo is AzureKaasInfo {
  return 'resourceGroups' in kaasInfo;
}

export function isGKEKaasInfo(kaasInfo: KaasInfo): kaasInfo is GKEKaasInfo {
  return 'cpu' in kaasInfo;
}

export function isEKSKaasInfo(kaasInfo: KaasInfo): kaasInfo is EKSKaasInfo {
  return 'amiTypes' in kaasInfo;
}
