import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';

import { EdgeStack } from '../types';

import { buildUrl } from './buildUrl';
import { queryKeys } from './query-keys';

interface StackFileResponse {
  StackFileContent: string;
}

export function useEdgeStackFile(id: EdgeStack['Id'], version?: number) {
  return useQuery(queryKeys.file(id, version), () =>
    getEdgeStackFile(id, version).catch(() => '')
  );
}

export async function getEdgeStackFile(id: EdgeStack['Id'], version?: number) {
  try {
    const { data } = await axios.get<StackFileResponse>(buildUrl(id, 'file'), {
      params: { version },
    });
    return data.StackFileContent;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
