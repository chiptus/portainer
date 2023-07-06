import { useMutation, useQuery, useQueryClient } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { success as notifySuccess } from '@/portainer/services/notifications';
import { withError } from '@/react-tools/react-query';

import {
  CreateCredentialPayload,
  Credential,
  UpdateCredentialPayload,
} from './types';

const queryKeys = {
  allCredentials: 'cloudcredentials',
  credential: (id: number) => ['cloudcredentials', `${id}`],
};

export async function createCredential(credential: CreateCredentialPayload) {
  try {
    await axios.post(buildUrl(), credential);
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to create credential');
  }
}

export async function getCloudCredentials() {
  try {
    const { data } = await axios.get<Credential[]>(buildUrl());
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to get credentials');
  }
}

export async function deleteCredential(credential: Credential) {
  try {
    await axios.delete<Credential[]>(buildUrl(credential.id));
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to delete credential');
  }
}

export async function getCloudCredential(id: number) {
  try {
    const { data } = await axios.get<Credential>(buildUrl(id));
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to get credentials');
  }
}

export async function updateCredential(
  credential: Partial<UpdateCredentialPayload>,
  id: number
) {
  try {
    const { data } = await axios.put(buildUrl(id), credential);
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to update credential');
  }
}

export function useUpdateCredentialMutation(id: number) {
  const queryClient = useQueryClient();

  return useMutation(
    ({ credential }: { credential: Partial<UpdateCredentialPayload> }) =>
      updateCredential(credential, id),
    {
      onSuccess: (_, data) => {
        notifySuccess(
          'Credentials updated successfully',
          data.credential.name || ''
        );
        return queryClient.invalidateQueries(queryKeys.credential(id));
      },
      meta: {
        error: {
          title: 'Failure',
          message: 'Unable to update credential',
        },
      },
    }
  );
}

export function useDeleteCredentialMutation() {
  const queryClient = useQueryClient();

  return useMutation(deleteCredential, {
    onSuccess: (_, credential) => {
      notifySuccess('Credentials deleted successfully', credential.name);
      return queryClient.invalidateQueries(queryKeys.allCredentials);
    },
    meta: {
      error: {
        title: 'Failure',
        message: 'Unable to delete credential',
      },
    },
  });
}

export function useCloudCredential(id: number) {
  return useQuery(queryKeys.credential(id), () => getCloudCredential(id), {
    cacheTime: 0, // don't cache to make sure the Use SSH key authentication is correctly set
    enabled: id >= 0,
    ...withError('Unable to retrieve cloud credential'),
  });
}

export function useCloudCredentials() {
  return useQuery(queryKeys.allCredentials, () => getCloudCredentials(), {
    staleTime: 20,
    meta: {
      error: {
        title: 'Failure',
        message: 'Unable to retrieve cloud credentials',
      },
    },
  });
}

export function useCreateCredentialMutation() {
  const queryClient = useQueryClient();

  return useMutation(createCredential, {
    onSuccess: (_, payload) => {
      notifySuccess('Credentials created successfully', payload.name);
      return queryClient.invalidateQueries(['cloudcredentials']);
    },
    meta: {
      error: {
        title: 'Failure',
        message: 'Unable to create credential',
      },
    },
  });
}

function buildUrl(credentialId?: number) {
  let url = 'cloudcredentials';
  if (credentialId) {
    url += `/${credentialId}`;
  }
  return url;
}
