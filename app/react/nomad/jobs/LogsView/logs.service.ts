import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';

export async function getTaskLogs(
  environmentId: EnvironmentId,
  namespace: string,
  jobID: string,
  allocationID: string,
  taskName: string,
  logType: string,
  refresh = false,
  offset = 0
) {
  try {
    const response = await axios.get(
      `/nomad/endpoints/${environmentId}/allocation/${allocationID}/logs`,
      {
        params: {
          jobID,
          taskName,
          namespace,
          refresh,
          logType,
          offset,
        },
      }
    );
    return response.data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
