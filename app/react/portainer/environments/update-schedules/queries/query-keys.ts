import { EnvironmentId } from '@/react/portainer/environments/types';

import { EdgeUpdateSchedule } from '../types';

export const queryKeys = {
  base: () => ['edge', 'update_schedules'] as const,
  list: (includeEdgeStacks?: boolean) =>
    [...queryKeys.base(), { includeEdgeStacks }] as const,
  item: (id: EdgeUpdateSchedule['id']) => [...queryKeys.base(), id] as const,
  activeSchedules: (environmentIds: EnvironmentId[]) =>
    [...queryKeys.base(), 'active', { environmentIds }] as const,
  supportedAgentVersions: () =>
    [...queryKeys.base(), 'agent_versions'] as const,
  previousVersions: (environmentIds?: EnvironmentId[]) =>
    [...queryKeys.base(), 'previous_versions', { environmentIds }] as const,
};
