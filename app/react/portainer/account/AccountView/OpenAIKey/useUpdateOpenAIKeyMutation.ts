import { useQueryClient, useMutation } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import {
  mutationOptions,
  withError,
  withInvalidate,
} from '@/react-tools/react-query';
import { userQueryKeys } from '@/portainer/users/queries/queryKeys';
import { buildUrl } from '@/portainer/users/user.service';
import { UserId } from '@/portainer/users/types';
import { useCurrentUser } from '@/react/hooks/useUser';

export interface UpdateUserOpenAIKeyPayload {
  ApiKey: string;
}

export function useUpdateUserOpenAIKeyMutation() {
  const queryClient = useQueryClient();

  const {
    user: { Id: userId },
  } = useCurrentUser();

  return useMutation(
    (query: UpdateUserOpenAIKeyPayload) => updateUserOpenAIKey(userId, query),
    mutationOptions(
      withInvalidate(queryClient, [userQueryKeys.base()]),
      withError('Unable to update OpenAI key')
    )
  );
}

async function updateUserOpenAIKey(
  userId: UserId,
  payload: UpdateUserOpenAIKeyPayload
) {
  try {
    await axios.put(buildUrl(userId, 'openai'), payload);
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to update OpenAI key');
  }
}
