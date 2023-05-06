import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';

import { EdgeStack } from '../types';

import { buildUrl } from './buildUrl';

export function useEdgeStacks() {
  return useQuery(['edge_stacks'], () => getEdgeStacks(), {
    ...withError('Failed loading Edge stack'),
  });
}

export async function getEdgeStacks() {
  try {
    const { data } = await axios.get<EdgeStack[]>(buildUrl());
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
