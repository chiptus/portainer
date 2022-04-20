import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/portainer/environments/types';

export async function deleteJob(
  environmentID: EnvironmentId,
  jobID: string,
  namespace: string
) {
  try {
    await axios.delete(`/nomad/jobs/${jobID}`, {
      params: { namespace, endpointId: environmentID },
    });
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
