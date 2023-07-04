import { useQuery } from 'react-query';

import { getMicroK8sInfo } from './service';

export function useMicroK8sOptions() {
  return useQuery(['microK8sClusterOptions'], () => getMicroK8sInfo());
}
