import { EnvironmentId } from '@/portainer/environments/types';

import { EdgeStack } from '../types';

export const queryKeys = {
  logsStatus: (edgeStackId: EdgeStack['Id'], environmentId: EnvironmentId) =>
    ['edge-stacks', edgeStackId, 'logs', environmentId] as const,
};
