import { useEffect } from 'react';
import { Box, Edit, Layers, Lock, Server, Shuffle } from 'lucide-react';
import { useQueryClient } from 'react-query';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { Authorized } from '@/react/hooks/useUser';
import Helm from '@/assets/ico/vendor/helm.svg?c';
import Route from '@/assets/ico/route.svg?c';
import { useEnvironmentDeploymentOptions } from '@/react/portainer/environments/queries/useEnvironment';

import { DashboardLink } from '../items/DashboardLink';
import { SidebarItem } from '../SidebarItem';
import { VolumesLink } from '../items/VolumesLink';
import { useSidebarState } from '../useSidebarState';

import { KubectlShellButton } from './KubectlShell';

interface Props {
  environmentId: EnvironmentId;
}

export function KubernetesSidebar({ environmentId }: Props) {
  const { isOpen } = useSidebarState();
  const queryClient = useQueryClient();

  const { data: deploymentOptions } =
    useEnvironmentDeploymentOptions(environmentId);
  const showCustomTemplates =
    deploymentOptions && !deploymentOptions?.hideWebEditor;

  // use an event listener to check when to invalidate the useEnvironmentDeploymentOptions query and update the side bar
  // this is a workaround for not being able to invalidate the query from the (angular based) settings and kube configure views
  useEffect(() => {
    document.addEventListener('portainer:deploymentOptionsUpdated', () => {
      // invalidate the query to update the sidebar
      queryClient.invalidateQueries([
        'environments',
        environmentId,
        'deploymentOptions',
      ]);
    });
    return () => {
      document.removeEventListener(
        'portainer:deploymentOptionsUpdated',
        () => {}
      );
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      {isOpen && (
        <div className="mb-3">
          <KubectlShellButton environmentId={environmentId} />
        </div>
      )}

      <DashboardLink
        environmentId={environmentId}
        platformPath="kubernetes"
        data-cy="k8sSidebar-dashboard"
      />

      {showCustomTemplates && (
        <SidebarItem
          to="kubernetes.templates.custom"
          params={{ endpointId: environmentId }}
          icon={Edit}
          label="Custom Templates"
          data-cy="k8sSidebar-customTemplates"
        />
      )}

      <SidebarItem
        to="kubernetes.resourcePools"
        params={{ endpointId: environmentId }}
        icon={Layers}
        label="Namespaces"
        data-cy="k8sSidebar-namespaces"
      />

      <Authorized
        authorizations="HelmInstallChart"
        environmentId={environmentId}
      >
        <SidebarItem
          to="kubernetes.templates.helm"
          params={{ endpointId: environmentId }}
          icon={Helm}
          label="Helm"
          data-cy="k8sSidebar-helm"
        />
      </Authorized>

      <SidebarItem
        to="kubernetes.applications"
        params={{ endpointId: environmentId }}
        icon={Box}
        label="Applications"
        data-cy="k8sSidebar-applications"
      />

      <SidebarItem
        to="kubernetes.services"
        params={{ endpointId: environmentId }}
        label="Services"
        data-cy="k8sSidebar-services"
        icon={Shuffle}
      />

      <SidebarItem
        to="kubernetes.ingresses"
        params={{ endpointId: environmentId }}
        label="Ingresses"
        data-cy="k8sSidebar-ingresses"
        icon={Route}
      />

      <SidebarItem
        to="kubernetes.configurations"
        params={{ endpointId: environmentId }}
        icon={Lock}
        label="ConfigMaps & Secrets"
        data-cy="k8sSidebar-configurations"
      />

      <VolumesLink
        environmentId={environmentId}
        platformPath="kubernetes"
        data-cy="k8sSidebar-volumes"
      />

      <SidebarItem
        label="Cluster"
        to="kubernetes.cluster"
        icon={Server}
        params={{ endpointId: environmentId }}
        data-cy="k8sSidebar-cluster"
      >
        <Authorized
          authorizations="K8sClusterSetupRW"
          adminOnlyCE
          environmentId={environmentId}
        >
          <SidebarItem
            to="kubernetes.cluster.setup"
            params={{ endpointId: environmentId }}
            label="Setup"
            data-cy="k8sSidebar-setup"
          />
        </Authorized>

        <Authorized
          authorizations="K8sClusterSetupRW"
          adminOnlyCE
          environmentId={environmentId}
        >
          <SidebarItem
            to="kubernetes.cluster.securityConstraint"
            params={{ endpointId: environmentId }}
            label="Security constraints"
            data-cy="k8sSidebar-securityConstraints"
          />
        </Authorized>

        <SidebarItem
          to="kubernetes.registries"
          params={{ endpointId: environmentId }}
          label="Registries"
          data-cy="k8sSidebar-registries"
        />
      </SidebarItem>
    </>
  );
}
