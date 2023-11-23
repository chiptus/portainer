import clsx from 'clsx';
import { DownloadCloud } from 'lucide-react';
import { useState } from 'react';

import {
  ContainerPlatform,
  useSystemInfo,
} from '@/react/portainer/system/useSystemInfo';
import { withEdition } from '@/react/portainer/feature-flags/withEdition';
import { withHideOnExtension } from '@/react/hooks/withHideOnExtension';
import { useCurrentUser } from '@/react/hooks/useUser';
import { useSystemVersion } from '@/react/portainer/system/useSystemVersion';
import { useUIState } from '@/react/hooks/useUIState';

import { Icon } from '@@/Icon';
import { Tooltip } from '@@/Tip/Tooltip';

import { useSidebarState } from '../useSidebarState';

import { UpdateDialog } from './UpdateDialog';
import styles from './UpdateBanner.module.css';

export const UpdateBannerWrapper = withHideOnExtension(
  withEdition(UpdateBanner, 'BE')
);

const enabledPlatforms: Array<ContainerPlatform> = [
  'Docker Standalone',
  'Docker Swarm',
  'Kubernetes',
];

function UpdateBanner() {
  const { isPureAdmin } = useCurrentUser();
  const { isOpen: isSidebarOpen } = useSidebarState();
  const systemInfoQuery = useSystemInfo();

  const [isOpen, setIsOpen] = useState(false);

  const uiStateStore = useUIState();
  const query = useSystemVersion();

  if (!query.data || !query.data.UpdateAvailable) {
    return null;
  }

  const { LatestVersion } = query.data;

  if (LatestVersion === uiStateStore.dismissedUpdateVersion) {
    return null;
  }

  if (!systemInfoQuery.data) {
    return null;
  }

  const systemInfo = systemInfoQuery.data;

  if (
    !enabledPlatforms.includes(systemInfo.platform) &&
    process.env.NODE_ENV !== 'development'
  ) {
    return null;
  }

  return (
    <div
      className={clsx(
        styles.root,
        'rounded border py-2',
        'bg-blue-11 th-dark:bg-gray-warm-11',
        'border-blue-9 th-dark:border-gray-warm-9'
      )}
    >
      {isSidebarOpen && (
        <div className={clsx(styles.dismissTitle, 'vertical-center')}>
          <Icon icon={DownloadCloud} mode="primary" size="md" />
          <span className="space-left">
            New version available {LatestVersion}
          </span>
        </div>
      )}

      <div className="inline-flex items-center">
        <button
          type="button"
          className={clsx(styles.dismissBtn)}
          onClick={() => onDismiss(LatestVersion)}
        >
          Dismiss
        </button>

        <button
          className={clsx(isPureAdmin ? styles.actions : styles.nonAdminAction)}
          disabled={!isPureAdmin}
          type="button"
          onClick={isPureAdmin ? handleClick : () => {}}
        >
          Update now
        </button>
        <div
          className={clsx(styles.updateTooltip, 'inline-flex', 'items-center')}
        >
          <Tooltip
            message={
              isPureAdmin
                ? 'This will effortlessly migrate you to the latest version of Portainer by updating this instance'
                : 'Admin users can effortlessly migrate to the latest version of Portainer by updating this instance'
            }
          />
        </div>
      </div>

      {isOpen && <UpdateDialog onDismiss={() => setIsOpen(false)} />}
    </div>
  );

  function handleClick() {
    setIsOpen(true);
  }

  function onDismiss(version: string) {
    uiStateStore.dismissUpdateVersion(version);
  }
}
