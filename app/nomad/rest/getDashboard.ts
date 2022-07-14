import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/portainer/environments/types';
import { Dashboard } from '@/nomad/types';

export async function getDashboard(environmentId: EnvironmentId) {
  try {
    const { data: dashboard } = await axios.get<Dashboard>(
      `/nomad/endpoints/${environmentId}/dashboard`,
      {
        params: {},
      }
    );
    return dashboard;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
