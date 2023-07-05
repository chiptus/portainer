import { useMutation, useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';
import { EnvironmentStatus } from '@/react/portainer/environments/types';

import { Option } from '@@/form-components/Input/Select';

interface AddonsResponse {
  microk8s: {
    running: boolean;
  };
  highAvailability: {
    enabled: boolean;
    nodes: {
      address: string;
      role: string;
    }[];
  };
  addons: {
    name: string;
    status: string;
    repository: string;
  }[];
  currentVersion: string;
  kubernetesVersions: Option<string>[];
}

async function getAddons(environmentID: number) {
  try {
    const { data } = await axios.get<AddonsResponse>(
      `cloud/endpoints/${environmentID}/addons`
    );
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to retrieve addons');
  }
}

async function upgradeCluster(environmentID: number, nextVersion: string) {
  try {
    const { data } = await axios.post<AddonsResponse>(
      `cloud/endpoints/${environmentID}/upgrade`,
      { nextVersion }
    );
    return data;
  } catch (err) {
    throw parseAxiosError(
      err as Error,
      'Unable to send upgrade cluster request'
    );
  }
}

async function updateAddons(
  environmentID: number,
  payload: { addons: string[] }
) {
  try {
    const { data } = await axios.post<AddonsResponse>(
      `cloud/endpoints/${environmentID}/addons`,
      payload
    );
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to update addons');
  }
}

export function useAddonsQuery<TSelect = AddonsResponse | null>(
  environmentID?: number,
  status?: number,
  select?: (info: AddonsResponse | null) => TSelect
) {
  return useQuery(
    ['clusterInfo', environmentID, 'addons'],
    () => (environmentID ? getAddons(environmentID) : null),
    {
      select,
      enabled: !!environmentID && status !== EnvironmentStatus.Error,
    }
  );
}

type UpdateAddOns = {
  environmentID: number;
  credentialID: number;
  payload: { addons: string[] };
};

type UpgradeRequest = {
  environmentID: number;
  nextVersion: string;
};

export function useUpdateAddonsMutation() {
  return useMutation(
    ({ environmentID, payload }: UpdateAddOns) =>
      updateAddons(environmentID, payload),
    withError('Failed to update addons')
  );
}

export function useUpgradeClusterMutation() {
  return useMutation(
    ({ environmentID, nextVersion }: UpgradeRequest) =>
      upgradeCluster(environmentID, nextVersion),
    withError('Failed to send upgrade cluster request')
  );
}
