import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useHideFormRedirect } from '@/react/kubernetes/cluster/deployment-options/useHideFormRedirect';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { useUnauthorizedRedirect } from '@/react/hooks/useUnauthorizedRedirect';

import { PageHeader } from '@@/PageHeader';

import { CreateNamespaceForm } from './CreateNamespaceForm';

export function CreateNamespaceView() {
  const environmentId = useEnvironmentId();
  useHideFormRedirect('kubernetes.resourcePools', {
    endpointId: environmentId,
  });

  useUnauthorizedRedirect(
    {
      authorizations: 'K8sResourcePoolsW',
      forceEnvironmentId: environmentId,
      adminOnlyCE: !isBE,
    },
    {
      to: 'kubernetes.resourcePools',
      params: {
        id: environmentId,
      },
    }
  );

  return (
    <div className="form-horizontal">
      <PageHeader
        title="Create a namespace"
        breadcrumbs="Create a namespace"
        reload
      />

      <div className="row">
        <div className="col-sm-12">
          <CreateNamespaceForm />
        </div>
      </div>
    </div>
  );
}
