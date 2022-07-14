import axios, { parseAxiosError } from 'Portainer/services/axios';
import { EnvironmentId } from 'Portainer/environments/types';

import { NomadEventsList } from '@/nomad/types';

export async function getTaskEvents(
  environmentId: EnvironmentId,
  allocationId: string,
  jobId: string,
  taskName: string,
  namespace: string
) {
  try {
    const ret = await axios.get<NomadEventsList>(
      `/nomad/endpoints/${environmentId}/allocation/${allocationId}/events`,
      {
        params: { jobId, taskName, namespace },
      }
    );
    return ret.data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
