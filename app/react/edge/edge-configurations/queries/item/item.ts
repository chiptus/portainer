import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';

import { EdgeConfiguration } from '../../types';
import { queryKeys } from '../query-keys';
import { buildUrl } from '../urls';

async function getItem(
  id: EdgeConfiguration['id']
): Promise<EdgeConfiguration> {
  try {
    const { data } = await axios.get<EdgeConfiguration>(buildUrl({ id }));
    return data;
  } catch (err) {
    throw parseAxiosError(
      err as Error,
      'Failed to retrieve edge configuration'
    );
  }
}

export function useEdgeConfiguration(id: EdgeConfiguration['id']) {
  return useQuery(queryKeys.item(id), async () => getItem(id), {
    ...withError('Failure retrieving configuration'),
  });
}
