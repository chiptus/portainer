import { useMutation } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';

type SSHKeyPair = {
  public: string;
  private: string;
};

export async function generateSSHKeyPair(passphrase?: string) {
  try {
    const { data } = await axios.post<SSHKeyPair>('/sshkeygen', {
      passphrase,
    });
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to generate SSH key pair');
  }
}

export function useGenerateSSHKeyMutation() {
  return useMutation(generateSSHKeyPair, {
    meta: {
      error: {
        title: 'Failure',
        message:
          'Unable to generate SSH key pair, try uploading one in the meantime.',
      },
    },
  });
}
