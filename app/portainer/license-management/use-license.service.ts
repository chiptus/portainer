import { useQuery } from 'react-query';

import { getLicenseInfo } from './license.service';
import { LicenseInfo } from './types';

export function useLicenseInfo() {
  const { isLoading, error, data: info } = useQuery<LicenseInfo, Error>(
    'licenseInfo',
    () => getLicenseInfo()
  );

  return { isLoading, error, info };
}
