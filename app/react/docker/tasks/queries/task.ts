import { useQuery } from 'react-query';

import { TaskId } from '@/react/docker/tasks/types';
import { getTask } from '@/react/docker/tasks/axios/getTask';
import { queryKeys } from '@/react/docker/tasks/queries/query-keys';
import { EnvironmentId } from '@/react/portainer/environments/types';

export function useTask(environmentId: EnvironmentId, taskId: TaskId) {
  return useQuery(
    queryKeys.task(environmentId, taskId),
    () => getTask(environmentId, taskId),
    {
      meta: {
        title: 'Failure',
        message: 'Unable to retrieve task',
      },
    }
  );
}
