import { useQuery, useMutation, useQueryClient } from 'react-query';

import {
  error as notifyError,
  success as notifySuccess,
} from '@/portainer/services/notifications';
// import { SettingsResponse } from '@/portainer/settings/settings.types';

import {
  publicSettings,
  updateSettings,
  getSettings,
  Settings,
} from './settings.service';
import { CloudSettingsAPIPayload } from './cloud/cloud.types';

export function usePublicSettings() {
  return useQuery(['settings', 'public'], () => publicSettings(), {
    onError: (err) => {
      notifyError('Failure', err as Error, 'Unable to retrieve settings');
    },
  });
}

export function useSettings<T = Settings>(select?: (settings: Settings) => T) {
  return useQuery(['settings'], getSettings, {
    select,
    onError: (err) => {
      notifyError('Failure', err as Error, 'Unable to retrieve settings');
    },
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();

  return useMutation(
    (payload: CloudSettingsAPIPayload) => updateSettings(payload),
    {
      onSuccess: () => {
        notifySuccess('Success', 'Settings updated successfully');
        queryClient.invalidateQueries(['settings']);
        // invalidate the cloud info too, incase the cloud api keys changed
        return queryClient.invalidateQueries(['cloud']);
      },
      onError: (err) => {
        notifyError('Failure', err as Error, 'Unable to update settings');
      },
    }
  );
}
