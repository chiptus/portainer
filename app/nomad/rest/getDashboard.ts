import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/portainer/environments/types';
import { Dashboard } from '@/nomad/types';

export async function getDashboard(environmentID: EnvironmentId) {
  try {
    const { data: dashboard } = await axios.get<Dashboard>(`/nomad/dashboard`, {
      params: { endpointId: environmentID },
    });
    return dashboard;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
