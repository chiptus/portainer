import axios, { parseAxiosError } from '@/portainer/services/axios';

import { EnvironmentId } from '../types';
import {
  RegistryId,
  Registry,
  RegistryAccess,
} from '../../registries/types/registry';

import { buildUrl } from './utils';

export async function updateEnvironmentRegistryAccess(
  environmentId: EnvironmentId,
  registryId: RegistryId,
  access: Partial<RegistryAccess>
) {
  try {
    await axios.put<void>(buildRegistryUrl(environmentId, registryId), access);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

export async function getEnvironmentRegistries(
  id: EnvironmentId,
  namespace: string
) {
  try {
    const { data } = await axios.get<Registry[]>(buildRegistryUrl(id), {
      params: { namespace },
    });
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

export async function getEnvironmentRegistry(
  endpointId: EnvironmentId,
  registryId: RegistryId
) {
  try {
    const { data } = await axios.get<Registry>(
      buildRegistryUrl(endpointId, registryId)
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

function buildRegistryUrl(id: EnvironmentId, registryId?: RegistryId) {
  let url = `${buildUrl(id)}/registries`;

  if (registryId) {
    url += `/${registryId}`;
  }

  return url;
}
