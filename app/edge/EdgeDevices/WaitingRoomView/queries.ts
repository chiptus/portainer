import { useMutation, useQueryClient } from 'react-query';

import { EnvironmentId } from '@/portainer/environments/types';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { promiseSequence } from '@/portainer/helpers/promise-utils';
import { useIntegratedLicenseInfo } from '@/portainer/license-management/use-license.service';
import { LicenseType } from '@/portainer/license-management/types';

export function useAssociateDeviceMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    (ids: EnvironmentId[]) =>
      promiseSequence(ids.map((id) => () => associateDevice(id))),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['environments']);
      },
      meta: {
        error: {
          title: 'Failure',
          message: 'Failed to associate devices',
        },
      },
    }
  );
}

async function associateDevice(environmentId: EnvironmentId) {
  try {
    await axios.post(`/endpoints/${environmentId}/edge/trust`);
  } catch (e) {
    throw parseAxiosError(e as Error, 'Failed to associate device');
  }
}

export function useLicenseOverused() {
  const integratedInfo = useIntegratedLicenseInfo();
  if (
    integratedInfo &&
    integratedInfo.usedNodes > integratedInfo.licenseInfo.nodes &&
    integratedInfo.licenseInfo.type === LicenseType.Essentials
  ) {
    return true;
  }
  return false;
}
