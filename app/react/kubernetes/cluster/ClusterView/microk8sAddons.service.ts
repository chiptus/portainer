import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';

export interface AddonsResponse {
  addons: string[];
}

export async function getAddons(environmentID: number, credentialID: number) {
  try {
    const { data } = await axios.get<AddonsResponse>('cloud/microk8s/addons', {
      params: { credentialID, environmentID },
    });
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to retrieve addons');
  }
}

export function useAddons<TSelect = AddonsResponse | null>(
  environmentID?: number,
  credentialID?: number,
  select?: (info: AddonsResponse | null) => TSelect
) {
  return useQuery(
    ['clusterInfo', environmentID, 'addons'],
    () =>
      environmentID && credentialID
        ? getAddons(environmentID, credentialID)
        : null,
    {
      select,
      enabled: !!environmentID && !!credentialID,
    }
  );
}
