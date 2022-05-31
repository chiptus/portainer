import { useQuery, useMutation, useQueryClient } from 'react-query';

import { notifyError } from '@/portainer/services/notifications';

import {
  publicSettings,
  updateSettings,
  getSettings,
} from './settings.service';
import { Settings } from './types';

export function usePublicSettings() {
  return useQuery(['settings', 'public'], () => publicSettings(), {
    onError: (err) => {
      notifyError('Failure', err as Error, 'Unable to retrieve settings');
    },
  });
}

export function useSettings<T = Settings>(
  select?: (settings: Settings) => T,
  enabled = true
) {
  return useQuery(['settings'], getSettings, {
    select,
    enabled,
    staleTime: 50,
    meta: {
      error: {
        title: 'Failure',
        message: 'Unable to retrieve settings',
      },
    },
  });
}

export function useUpdateSettingsMutation() {
  const queryClient = useQueryClient();

  return useMutation((payload: Partial<Settings>) => updateSettings(payload), {
    onSuccess: async () => {
      await queryClient.invalidateQueries(['settings']);
      // invalidate the cloud info too, incase the cloud api keys changed
      return queryClient.invalidateQueries(['cloud']);
    },
    meta: {
      error: {
        title: 'Failure',
        message: 'Unable to update settings',
      },
    },
  });
}
