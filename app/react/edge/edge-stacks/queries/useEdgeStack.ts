import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';

import { EdgeStack } from '../types';

import { buildUrl } from './buildUrl';

export function useEdgeStack(id: EdgeStack['Id']) {
  return useQuery(['edge_stacks', id], () => getEdgeStack(id), {
    ...withError('Failed loading Edge stack'),
  });
}

export async function getEdgeStack(id: EdgeStack['Id']) {
  try {
    const { data } = await axios.get<EdgeStack>(buildUrl(id));
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
