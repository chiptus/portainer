import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/portainer/environments/types';
import { Job } from '@/nomad/types';

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

export async function listJobs(environmentID: EnvironmentId) {
  try {
    const { data: jobs } = await axios.get<Job[]>(`/nomad/jobs`, {
      params: { endpointId: environmentID },
    });
    return jobs;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
