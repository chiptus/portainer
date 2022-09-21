import { AxiosError } from 'axios';
import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';

interface CheckPayload {
  repository: string;
  username?: string;
  password?: string;
}

export function useCheckRepo(payload: CheckPayload, enabled: boolean) {
  return useQuery<string[], Error>(
    ['git_repo_valid', { payload }],
    () => checkRepo(payload),
    {
      enabled,
      retry: false,
    }
  );
}

export async function checkRepo(payload: CheckPayload) {
  try {
    const { data } = await axios.post<string[]>('/gitops/repo/refs', payload);
    return data;
  } catch (error) {
    throw parseAxiosError(error as Error, '', (axiosError: AxiosError) => {
      let details = axiosError.response?.data.details;

      // If no credentials were provided alter error from git to indicate repository is not found or is private
      if (
        !(payload.username && payload.password) &&
        details ===
          'Authentication failed, please ensure that the git credentials are correct.'
      ) {
        details =
          'Git repository could not be found or is private, please ensure that the URL is correct or credentials are provided.';
      }

      const error = new Error(details);
      return { error, details };
    });
  }
}
