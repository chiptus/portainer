import _ from 'lodash';

import axios from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import PortainerError from '@/portainer/error';
import { urlBuilder } from '@/react/docker/tasks/axios/urlBuilder';
import { TaskId, TaskLogsParams } from '@/react/docker/tasks/types';

export async function getTaskLogs(
  environmentId: EnvironmentId,
  taskId: TaskId,
  params?: TaskLogsParams
): Promise<string> {
  try {
    const { data } = await axios.get<string>(
      urlBuilder(environmentId, taskId, 'logs'),
      {
        params: _.pickBy(params),
      }
    );

    return data;
  } catch (e) {
    throw new PortainerError('Unable to get task logs', e as Error);
  }
}
