import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';

export interface AddonsResponse {
  addons: string[];
}

export async function getAddons(environmentID: number, credentialId: number) {
  try {
    const { data } = await axios.get<AddonsResponse>('cloud/microk8s/info', {
      params: { credentialId, environmentID },
    });
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to retrieve addons');
  }
}

export function useAddons<TSelect = AddonsResponse | null>(
  environmentID?: number,
  credentialId?: number,
  select?: (info: AddonsResponse | null) => TSelect
) {
  return useQuery(
    ['clusterInfo', environmentID, 'addons'],
    () =>
      environmentID && credentialId
        ? getAddons(environmentID, credentialId)
        : null,
    {
      select,
      enabled: !!environmentID && !!credentialId,
    }
  );
}
