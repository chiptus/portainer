import { useEnvironment } from '@/portainer/environments/queries';

import { useEnvironmentId } from './useEnvironmentId';

export function useCurrentEnvironment() {
  const id = useEnvironmentId();
  return useEnvironment(id);
}
