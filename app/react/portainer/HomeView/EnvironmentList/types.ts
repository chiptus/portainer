import { TagId } from '@/portainer/tags/types';
import { EnvironmentGroupId } from '@/react/portainer/environments/environment-groups/types';
import {
  PlatformType,
  EnvironmentStatus,
} from '@/react/portainer/environments/types';

export interface Filter<T = number> {
  value: T;
  label: string;
}

export enum ConnectionType {
  API,
  Agent,
  EdgeAgent,
  EdgeDevice,
}

export interface Filters {
  platformTypes: Array<PlatformType>;
  connectionTypes: Array<ConnectionType>;
  status: Array<EnvironmentStatus>;
  tagIds?: Array<TagId>;
  groupIds: Array<EnvironmentGroupId>;
  agentVersions: Array<string>;
  sort?: string;
  sortDesc: boolean;
}
