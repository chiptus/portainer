import { EdgeGroup } from '@/react/edge/edge-groups/types';
import { RegistryId } from '@/react/portainer/registries/types/registry';

import { ScheduleType } from '../types';

export interface FormValues {
  name: string;
  groupIds: EdgeGroup['Id'][];
  type: ScheduleType;
  version: string;
  scheduledTime: string;
  registryId: RegistryId;
}
