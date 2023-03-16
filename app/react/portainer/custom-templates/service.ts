import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';

import { CustomTemplate } from './types';

export async function getCustomTemplates() {
  try {
    const { data } = await axios.get<CustomTemplate[]>(`custom_templates`);
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to get custom templates');
  }
}

export function useCustomTemplates() {
  return useQuery('customtemplates', () => getCustomTemplates(), {
    staleTime: 20,
    meta: {
      error: {
        title: 'Failure',
        message: 'Unable to retrieve custom templates',
      },
    },
  });
}
