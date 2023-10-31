import { useCurrentStateAndParams, useRouter } from '@uirouter/react';
import { useEffect } from 'react';
import { X, Slash, History } from 'lucide-react';
import clsx from 'clsx';
import { useStore } from 'zustand';

import {
  PlatformType,
  EnvironmentId,
  Environment,
} from '@/react/portainer/environments/types';
import {
  getPlatformType,
  isEdgeAsync,
} from '@/react/portainer/environments/utils';
import { useEnvironment } from '@/react/portainer/environments/queries/useEnvironment';
import { environmentStore } from '@/react/hooks/current-environment-store';

import { Icon } from '@@/Icon';

import { getPlatformIcon } from '../portainer/environments/utils/get-platform-icon';

import styles from './EnvironmentSidebar.module.css';
import { AzureSidebar } from './AzureSidebar';
import { DockerSidebar } from './DockerSidebar';
import { KubernetesSidebar } from './KubernetesSidebar';
import { SidebarSection, SidebarSectionTitle } from './SidebarSection';
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
  const isBrowsingSnapshot = isEdgeAsync(environment);

  return (
    <>
      {isBrowsingSnapshot && (
        <div
          className={clsx(
            styles.root,
            '!mb-0 rounded-t border-x border-b-0 border-t border-dotted py-2'
          )}
        >
          <SnapshotBrowseSection />
        </div>
      )}
      <div
        className={clsx(
          styles.root,
          isBrowsingSnapshot ? '!mt-0 rounded-b' : 'rounded',
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
              <span className="text-sm text-gray-6">None selected</span>
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
    };

    return sidebar[platform];
  }
}

function useCurrentEnvironment() {
  const { params } = useCurrentStateAndParams();
  const router = useRouter();

  const envStore = useStore(environmentStore);
  const { setEnvironmentId } = envStore;
  useEffect(() => {
    const envIdParam = params.environmentId || params.endpointId;
    const environmentId = parseInt(envIdParam, 10);

    if (envIdParam && !Number.isNaN(environmentId)) {
      setEnvironmentId(environmentId);
    }
  }, [setEnvironmentId, params.endpointId, params.environmentId]);

  return { query: useEnvironment(envStore.environmentId), clearEnvironment };

  function clearEnvironment() {
    if (params.endpointId || params.environmentId) {
      router.stateService.go('portainer.home');
    }

    envStore.clear();
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
        <div className="-ml-3 flex w-8 justify-center">
          <span className="label-warning h-2.5 w-2.5 rounded-full" />
        </div>
      </SidebarSectionTitle>
    );
  }
  return (
    <SidebarSectionTitle>
      <div className="flex items-center">
        <span className="label-warning ml-1 mr-3 h-2.5 w-2.5 rounded-full" />
        <span className="overflow-hidden text-ellipsis whitespace-nowrap text-white">
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
      <div className="-ml-3 flex w-8 justify-center" title={environment.Name}>
        <EnvironmentIcon className="text-2xl" />
      </div>
    );
  }

  return (
    <div className="flex items-center">
      <EnvironmentIcon className="mr-3 text-2xl" />
      <span className="overflow-hidden text-ellipsis whitespace-nowrap text-white">
        {environment.Name}
      </span>

      <button
        title="Clear environment"
        type="button"
        onClick={onClear}
        className={clsx(
          styles.closeBtn,
          'ml-auto mr-2 flex h-5 w-5 items-center justify-center rounded border-0 p-1 text-sm text-gray-5 transition-colors duration-200 hover:text-white be:text-gray-6 be:hover:text-white'
        )}
      >
        <X />
      </button>
    </div>
  );
}
