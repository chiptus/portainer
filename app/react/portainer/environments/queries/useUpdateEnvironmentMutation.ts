import { useQueryClient, useMutation } from 'react-query';

import {
  mutationOptions,
  withError,
  withInvalidate,
} from '@/react-tools/react-query';
import {
  EnvironmentId,
  EnvironmentStatusMessage,
} from '@/react/portainer/environments/types';

import { updateEndpointStatusMessage } from '../environment.service';

export function useUpdateEndpointStatusMessageMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    ({
      environmentId,
      payload,
    }: {
      environmentId: EnvironmentId;
      payload: {
        IsSetStatusMessage: boolean;
        StatusMessage: EnvironmentStatusMessage;
      };
    }) => updateEndpointStatusMessage(environmentId, payload),
    mutationOptions(
      withError('Failed to update environment status message'),
      withInvalidate(queryClient, [['environments']])
    )
  );
}
