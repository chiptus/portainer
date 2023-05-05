import { useQuery } from 'react-query';

import { error as notifyError } from '@/portainer/services/notifications';
import { useNodesCount } from '@/react/portainer/system/useNodesCount';

import { getLicenseInfo, getLicenses } from './license.service';
import { LicenseType, LicenseInfo, License } from './types';

export const queryKeys = {
  base: () => ['licenses'] as const,
  licenseInfo: () => [...queryKeys.base(), 'info'] as const,
};

// returns the aggregated license info (expires at has the latest date, nodes has the sum of all nodes)
export function useLicenseInfo() {
  const { isLoading, data: info } = useQuery<LicenseInfo, Error>(
    queryKeys.licenseInfo(),
    () => getLicenseInfo(),
    {
      onError(error) {
        notifyError('Failure', error as Error, 'Failed to get license info');
      },
    }
  );

  return { isLoading, info };
}

// returns a use query hook for the list of portainer licenses
export function useLicensesQuery() {
  return useQuery<License[], Error>(queryKeys.base(), () => getLicenses(), {
    onError(error) {
      notifyError('Failure', error as Error, 'Failed to get licenses');
    },
  });
}

export function useIntegratedLicenseInfo() {
  const { isLoading: isLoadingNodes, data: nodesCount = 0 } = useNodesCount();

  const { isLoading: isLoadingLicense, info } = useLicenseInfo();
  if (
    isLoadingLicense ||
    isLoadingNodes ||
    !info ||
    info.type === LicenseType.Trial
  ) {
    return null;
  }

  return { licenseInfo: info as LicenseInfo, usedNodes: nodesCount };
}
