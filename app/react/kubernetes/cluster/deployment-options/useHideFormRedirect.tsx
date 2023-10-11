import { useRouter, StateOrName, RawParams } from '@uirouter/react';
import { useEffect } from 'react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironmentDeploymentOptions } from '@/react/portainer/environments/queries/useEnvironment';

export function useHideFormRedirect(route: StateOrName, params: RawParams) {
  const router = useRouter();
  const environmentId = useEnvironmentId();
  const { data: deploymentOptions } =
    useEnvironmentDeploymentOptions(environmentId);
  useEffect(() => {
    if (deploymentOptions?.hideAddWithForm) {
      router.stateService.go(route, params);
    }
  }, [deploymentOptions?.hideAddWithForm, params, route, router.stateService]);
}
