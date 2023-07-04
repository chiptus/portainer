import { useMutation } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';

type TestSSHConnectionPayload = {
  nodeIPs: string[];
  credentialID: number;
};

export type TestSSHConnectionResponse = {
  address: string;
  error?: string;
}[];

export function useTestSSHConnection() {
  return useMutation(
    (payload: TestSSHConnectionPayload) => testSSHConnection(payload),
    withError('Unable to test SSH connection')
  );
}

async function testSSHConnection(payload: TestSSHConnectionPayload) {
  try {
    const { data } = await axios.post<TestSSHConnectionResponse>(
      `/cloud/testssh`,
      payload
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
