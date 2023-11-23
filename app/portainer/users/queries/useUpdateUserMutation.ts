import { useQueryClient, useMutation } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import {
  mutationOptions,
  withError,
  withInvalidate,
} from '@/react-tools/react-query';
import { userQueryKeys } from '@/portainer/users/queries/queryKeys';
import { buildUrl } from '@/portainer/users/buildUrl';
import { Role, UserId } from '@/portainer/users/types';

export interface UpdateUserPayload {
  username?: string;
  password?: string;
  newPassword?: string;
  role?: Role;
}

type Update = {
  userId: UserId;
  payload: UpdateUserPayload;
};

export function useUpdateUserMutation() {
  const queryClient = useQueryClient();
  return useMutation(
    updateUser,
    mutationOptions(
      withInvalidate(queryClient, [userQueryKeys.base()]),
      withError('Unable to update user')
    )
  );
}

export async function updateUser({ userId, payload }: Update) {
  try {
    await axios.put(buildUrl(userId), payload);
  } catch (err) {
    throw parseAxiosError(err, 'Unable to update user');
  }
}
