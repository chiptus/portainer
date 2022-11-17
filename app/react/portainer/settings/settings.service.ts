import { PublicSettingsViewModel } from '@/portainer/models/settings';
import axios, { parseAxiosError } from '@/portainer/services/axios';

import { PublicSettingsResponse, DefaultRegistry, Settings } from './types';

export interface GlobalDeploymentOptions {
  perEnvOverride: boolean;
  hideAddWithForm: boolean;
  hideWebEditor: boolean;
  hideFileUpload: boolean;
}

export async function getPublicSettings() {
  try {
    const { data } = await axios.get<PublicSettingsResponse>(
      buildUrl('public')
    );
    return new PublicSettingsViewModel(data);
  } catch (e) {
    throw parseAxiosError(
      e as Error,
      'Unable to retrieve application settings'
    );
  }
}

export async function getGlobalDeploymentOptions() {
  try {
    const { data } = await axios.get<PublicSettingsResponse>(
      buildUrl('public')
    );
    return new PublicSettingsViewModel(data)
      .GlobalDeploymentOptions as GlobalDeploymentOptions;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve deployment options');
  }
}

export async function getSettings() {
  try {
    const { data } = await axios.get<Settings>(buildUrl());
    return data;
  } catch (e) {
    throw parseAxiosError(
      e as Error,
      'Unable to retrieve application settings'
    );
  }
}

export async function updateSettings(settings: Partial<Settings>) {
  try {
    await axios.put(buildUrl(), settings);
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to update application settings');
  }
}

export async function updateDefaultRegistry(
  defaultRegistry: Partial<DefaultRegistry>
) {
  try {
    await axios.put(buildUrl('default_registry'), defaultRegistry);
  } catch (e) {
    throw parseAxiosError(
      e as Error,
      'Unable to update default registry settings'
    );
  }
}

function buildUrl(subResource?: string, action?: string) {
  let url = 'settings';
  if (subResource) {
    url += `/${subResource}`;
  }

  if (action) {
    url += `/${action}`;
  }

  return url;
}
