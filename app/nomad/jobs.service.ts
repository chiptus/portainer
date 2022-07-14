import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/portainer/environments/types';
import { Job } from '@/nomad/types';

export async function deleteJob(
  environmentId: EnvironmentId,
  jobId: string,
  namespace: string
) {
  try {
    await axios.delete(`/nomad/endpoints/${environmentId}/jobs/${jobId}`, {
      params: { namespace },
    });
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

export async function listJobs(environmentId: EnvironmentId) {
  try {
    const { data: jobs } = await axios.get<Job[]>(
      `/nomad/endpoints/${environmentId}/jobs`,
      {
        params: {},
      }
    );
    return jobs;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
