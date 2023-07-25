import { useQueryClient, useMutation } from 'react-query';

import { promiseSequence } from '@/portainer/helpers/promise-utils';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import {
  mutationOptions,
  withError,
  withInvalidate,
} from '@/react-tools/react-query';

import { EdgeConfiguration } from '../types';

import { buildUrl } from './urls';
import { queryKeys } from './query-keys';

export function useRemoveMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    (configurations: EdgeConfiguration[]) =>
      promiseSequence(
        configurations.map(
          (configuration) => () => deleteConfiguration(configuration.id)
        )
      ),

    mutationOptions(
      withInvalidate(queryClient, [queryKeys.base()]),
      withError()
    )
  );
}

async function deleteConfiguration(id: EdgeConfiguration['id']) {
  try {
    const { data } = await axios.delete<EdgeConfiguration[]>(buildUrl({ id }));
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Failed to delete edge configuration');
  }
}
