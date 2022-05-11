import { useMutation, useQuery, useQueryClient } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { success as notifySuccess } from '@/portainer/services/notifications';

import {
  CreateCredentialPayload,
  Credential,
  UpdateCredentialPayload,
} from './types';

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
    throw parseAxiosError(e as Error, 'Unable to get credential');
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

export function useUpdateCredentialMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    ({
      credential,
      id,
    }: {
      credential: Partial<UpdateCredentialPayload>;
      id: number;
    }) => updateCredential(credential, id),
    {
      onSuccess: (_, data) => {
        notifySuccess('Credential updated successfully', data.credential.name);
        return queryClient.invalidateQueries(['cloudcredentials']);
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
      notifySuccess('Credential deleted successfully', credential.name);
      return queryClient.invalidateQueries(['cloudcredentials']);
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
  return useQuery(['cloudcredentials', id], () => getCloudCredential(id), {
    meta: {
      error: {
        title: 'Failure',
        message: 'Unable to retrieve cloud credential',
      },
    },
  });
}

export function useCloudCredentials() {
  return useQuery('cloudcredentials', () => getCloudCredentials(), {
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
