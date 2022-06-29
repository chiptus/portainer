import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/portainer/environments/types';
import { Leader } from '@/nomad/types';

export async function getLeader(environmentID: EnvironmentId) {
  try {
    const { data } = await axios.get<Leader>(`/nomad/leader`, {
      params: { endpointId: environmentID },
    });
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
