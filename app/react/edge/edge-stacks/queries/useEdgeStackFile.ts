import axios, { parseAxiosError } from '@/portainer/services/axios';

import { EdgeStack } from '../types';

import { buildUrl } from './buildUrl';

interface StackFileResponse {
  StackFileContent: string;
}

export async function getEdgeStackFile(id?: EdgeStack['Id'], version?: number) {
  if (!id) {
    return null;
  }

  try {
    const { data } = await axios.get<StackFileResponse>(buildUrl(id, 'file'), {
      params: { version },
    });
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
