import { useMutation, useQueryClient } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError, withInvalidate } from '@/react-tools/react-query';

import { FormValues } from '../../common/types';
import { EdgeConfiguration } from '../../types';
import { queryKeys } from '../query-keys';
import { buildUrl } from '../urls';

interface Update {
  id: EdgeConfiguration['id'];
  values: Partial<FormValues>;
}

async function update({ id, values }: Update) {
  try {
    const { data } = await axios.put<EdgeConfiguration>(
      buildUrl({ id }),
      values
    );

    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Failed to update edge configuration');
  }
}

export function useUpdateMutation() {
  const queryClient = useQueryClient();
  return useMutation(update, {
    ...withInvalidate(queryClient, [queryKeys.base()]),
    ...withError(),
  });
}
