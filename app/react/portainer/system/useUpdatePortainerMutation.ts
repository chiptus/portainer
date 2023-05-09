import { useMutation } from 'react-query';

import axios, {
  isAxiosError,
  parseAxiosError,
} from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';

import { buildUrl } from './build-url';

export function useUpdatePortainerMutation() {
  return useMutation(updatePortainer, {
    ...withError('Unable to update Portainer'),
  });
}

async function updatePortainer() {
  try {
    await axios.post(buildUrl('update'));
  } catch (error) {
    if (!isAxiosError(error)) {
      throw error;
    }

    // if the server is disconnected, then everything went well
    if (!error.response || !error.response.status) {
      return;
    }

    throw parseAxiosError(error);
  }
}
