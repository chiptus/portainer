import axios, { parseAxiosError } from 'Portainer/services/axios';
import { EnvironmentId } from 'Portainer/environments/types';

import { NomadEventsList } from '@/nomad/types';

export async function getTaskEvents(
  environmentID: EnvironmentId,
  allocationID: string,
  jobID: string,
  taskName: string,
  namespace: string
) {
  try {
    const ret = await axios.get<NomadEventsList>(
      `/nomad/allocation/${allocationID}/events`,
      {
        params: { jobID, taskName, namespace, endpointId: environmentID },
      }
    );
    return ret.data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
