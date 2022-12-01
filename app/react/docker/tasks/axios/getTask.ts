import axios from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import PortainerError from '@/portainer/error';
import { urlBuilder } from '@/react/docker/tasks/axios/urlBuilder';
import { DockerTaskResponse, TaskId } from '@/react/docker/tasks/types';

export async function getTask(
  environmentId: EnvironmentId,
  taskId: TaskId
): Promise<DockerTaskResponse> {
  try {
    const { data } = await axios.get(urlBuilder(environmentId, taskId));

    return data;
  } catch (e) {
    throw new PortainerError('Unable to get task', e as Error);
  }
}
