import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';

import { CustomTemplate, CustomTemplateFileContent } from './types';

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

export async function getCustomTemplateFileContent(id: number) {
  try {
    const {
      data: { FileContent },
    } = await axios.get<CustomTemplateFileContent>(
      `custom_templates/${id}/file`
    );
    return FileContent;
  } catch (e) {
    throw parseAxiosError(
      e as Error,
      'Unable to get custom template file content'
    );
  }
}
