import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { Stack } from '@/react/docker/stacks/types';

import { buildStackUrl } from './buildUrl';

export function useStacks() {
  return useQuery(['stacks'], () => getStacks(), {
    ...withError('Failed loading stack'),
  });
}

export async function getStacks() {
  try {
    const { data } = await axios.get<Stack[]>(buildStackUrl());
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
