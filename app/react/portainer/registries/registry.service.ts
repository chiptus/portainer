import axios, { parseAxiosError } from '@/portainer/services/axios';

import { Catalog, Repository, Registry } from './types/registry';

export async function getRegistries() {
  try {
    const { data } = await axios.get<Registry[]>('/registries');
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve registries');
  }
}

export async function listRegistryCatalogs(registryId: number) {
  try {
    const { data } = await axios.get<Catalog>(
      `/registries/${registryId}/v2/_catalog`
    );
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Failed to get catelog of regisry');
  }
}

export async function listRegistryCatalogsRepository(
  registryId: number,
  repositoryName: string
) {
  try {
    const { data } = await axios.get<Repository>(
      `/registries/${registryId}/v2/${repositoryName}/tags/list`,
      {}
    );
    return data;
  } catch (err) {
    throw parseAxiosError(
      err as Error,
      'Failed to get catelog repository of regisry'
    );
  }
}
