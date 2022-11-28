import { useCurrentStateAndParams, useRouter } from '@uirouter/react';
import { useEffect } from 'react';
import { X, Slash, History } from 'lucide-react';
import clsx from 'clsx';
import angular from 'angular';

import {
  PlatformType,
  EnvironmentId,
  Environment,
} from '@/react/portainer/environments/types';
import {
  getPlatformType,
  isEdgeDeviceAsync,
} from '@/react/portainer/environments/utils';
import { useEnvironment } from '@/react/portainer/environments/queries/useEnvironment';
import { useLocalStorage } from '@/react/hooks/useLocalStorage';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { EndpointProviderInterface } from '@/portainer/services/endpointProvider';

import { Icon } from '@@/Icon';

import { getPlatformIcon } from '../portainer/environments/utils/get-platform-icon';

import styles from './EnvironmentSidebar.module.css';
import { AzureSidebar } from './AzureSidebar';
import { DockerSidebar } from './DockerSidebar';
import { KubernetesSidebar } from './KubernetesSidebar';
import { SidebarSection, SidebarSectionTitle } from './SidebarSection';
import { NomadSidebar } from './NomadSidebar';
import { useSidebarState } from './useSidebarState';
import { EdgeDeviceAsyncSidebar } from './EdgeDeviceAsyncSidebar';

export function EnvironmentSidebar() {
  const { query: currentEnvironmentQuery, clearEnvironment } =
    useCurrentEnvironment();
  const environment = currentEnvironmentQuery.data;

  const { isOpen } = useSidebarState();

  if (!isOpen && !environment) {
    return null;
  }
  const isBrowsingSnapshot = isEdgeDeviceAsync(environment);

  return (
    <>
      {isBrowsingSnapshot && (
        <div
          className={clsx(
            styles.root,
            'rounded-t border-x border-t border-b-0 border-dotted py-2 !mb-0'
          )}
        >
          <SnapshotBrowseSection />
        </div>
      )}
      <div
        className={clsx(
          styles.root,
          isBrowsingSnapshot ? 'rounded-b !mt-0' : 'rounded',
          'border border-dotted py-2'
        )}
      >
        {environment ? (
          <Content
            environment={environment}
            onClear={clearEnvironment}
            isBrowsingSnapshot={isBrowsingSnapshot}
          />
        ) : (
          <SidebarSectionTitle>
            <div className="flex items-center gap-1">
              <span>Environment:</span>
              <Icon icon={Slash} className="text-xl !text-gray-6" />
              <span className="text-gray-6 text-sm">None selected</span>
            </div>
          </SidebarSectionTitle>
        )}
      </div>
    </>
  );
}

interface ContentProps {
  environment: Environment;
  onClear: () => void;
  isBrowsingSnapshot: boolean;
}

function Content({ environment, onClear, isBrowsingSnapshot }: ContentProps) {
  const platform = getPlatformType(environment.Type);
  const Sidebar = getSidebar(platform, isBrowsingSnapshot);

  return (
    <SidebarSection
      title={<Title environment={environment} onClear={onClear} />}
      aria-label={environment.Name}
      showTitleWhenClosed
    >
      <div className="mt-2">
        {Sidebar && (
          <Sidebar environmentId={environment.Id} environment={environment} />
        )}
      </div>
    </SidebarSection>
  );

  function getSidebar(platform: PlatformType, isBrowsingSnapshot: boolean) {
    const sidebar: {
      [key in PlatformType]: React.ComponentType<{
        environmentId: EnvironmentId;
        environment: Environment;
      }> | null;
    } = {
      [PlatformType.Azure]: AzureSidebar,
      [PlatformType.Docker]: isBrowsingSnapshot
        ? EdgeDeviceAsyncSidebar
        : DockerSidebar,
      [PlatformType.Kubernetes]: KubernetesSidebar,
      [PlatformType.Nomad]: isBE ? NomadSidebar : null,
    };

    return sidebar[platform];
  }
}

function useCurrentEnvironment() {
  const { params } = useCurrentStateAndParams();
  const router = useRouter();
  const [environmentId, setEnvironmentId] = useLocalStorage<
    EnvironmentId | undefined
  >('environmentId', undefined, sessionStorage);

  useEffect(() => {
    const envIdParam = params.environmentId || params.endpointId;
    const environmentId = parseInt(envIdParam, 10);

    if (envIdParam && !Number.isNaN(environmentId)) {
      setEnvironmentId(environmentId);
    }
  }, [params.endpointId, params.environmentId, setEnvironmentId]);

  return { query: useEnvironment(environmentId), clearEnvironment };

  function clearEnvironment() {
    const $injector = angular.element(document).injector();
    $injector.invoke(
      /* @ngInject */ (EndpointProvider: EndpointProviderInterface) => {
        EndpointProvider.setCurrentEndpoint(null);
        if (!params.endpointId) {
          document.title = 'Portainer';
        }
      }
    );

    if (params.endpointId) {
      router.stateService.go('portainer.home');
    }

    setEnvironmentId(undefined);
  }
}

interface TitleProps {
  environment: Environment;
  onClear(): void;
}

function SnapshotBrowseSection() {
  const { isOpen } = useSidebarState();

  if (!isOpen) {
    return (
      <SidebarSectionTitle showWhenClosed>
        <div className="w-8 flex justify-center -ml-3">
          <span className="w-2.5 h-2.5 rounded-full label-warning" />
        </div>
      </SidebarSectionTitle>
    );
  }
  return (
    <SidebarSectionTitle>
      <div className="flex items-center">
        <span className="w-2.5 h-2.5 rounded-full label-warning ml-1 mr-3" />
        <span className="text-white text-ellipsis overflow-hidden whitespace-nowrap">
          Browsing snapshot
        </span>

        <Icon
          icon={History}
          className="!ml-auto !mr-3 text-gray-5 be:text-gray-6"
        />
      </div>
    </SidebarSectionTitle>
  );
}

function Title({ environment, onClear }: TitleProps) {
  const { isOpen } = useSidebarState();

  const EnvironmentIcon = getPlatformIcon(environment.Type);

  if (!isOpen) {
    return (
      <div className="w-8 flex justify-center -ml-3" title={environment.Name}>
        <EnvironmentIcon className="text-2xl" />
      </div>
    );
  }

  return (
    <div className="flex items-center">
      <EnvironmentIcon className="text-2xl mr-3" />
      <span className="text-white text-ellipsis overflow-hidden whitespace-nowrap">
        {environment.Name}
      </span>

      <button
        title="Clear environment"
        type="button"
        onClick={onClear}
        className={clsx(
          styles.closeBtn,
          'flex items-center justify-center transition-colors duration-200 rounded border-0 text-sm h-5 w-5 p-1 ml-auto mr-2 text-gray-5 be:text-gray-6 hover:text-white be:hover:text-white'
        )}
      >
        <X />
      </button>
    </div>
  );
}
