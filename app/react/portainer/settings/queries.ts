import { useQuery, useMutation, useQueryClient } from 'react-query';

import {
  mutationOptions,
  withError,
  withInvalidate,
} from '@/react-tools/react-query';

import {
  updateSettings,
  getSettings,
  getPublicSettings,
  updateDefaultRegistry,
} from './settings.service';
import { DefaultRegistry, PublicSettingsResponse, Settings } from './types';

export function usePublicSettings<T = PublicSettingsResponse>({
  enabled,
  select,
  onSuccess,
}: {
  select?: (settings: PublicSettingsResponse) => T;
  enabled?: boolean;
  onSuccess?: (data: T) => void;
} = {}) {
  return useQuery(['settings', 'public'], getPublicSettings, {
    select,
    ...withError('Unable to retrieve public settings'),
    enabled,
    onSuccess,
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
    ...withError('Unable to retrieve settings'),
  });
}

export function useUpdateSettingsMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    updateSettings,
    mutationOptions(
      withInvalidate(queryClient, [['settings'], ['cloud']]),
      withError('Unable to update settings')
    )
  );
}

export function useUpdateDefaultRegistrySettingsMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    (payload: Partial<DefaultRegistry>) => updateDefaultRegistry(payload),
    mutationOptions(
      withInvalidate(queryClient, [['settings']]),
      withError('Unable to update default registry settings')
    )
  );
}
