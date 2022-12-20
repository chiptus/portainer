import { useRouter } from '@uirouter/react';
import { useState } from 'react';
import { useStore } from 'zustand';

import { Environment } from '@/react/portainer/environments/types';
import { snapshotEndpoints } from '@/react/portainer/environments/environment.service';
import { isEdgeEnvironment } from '@/react/portainer/environments/utils';
import * as notifications from '@/portainer/services/notifications';
import { confirmAsync } from '@/portainer/services/modal.service/confirm';
import { buildTitle } from '@/portainer/services/modal.service/utils';
import { DockerSnapshotPanel } from '@/react/edge/components/SnapshotPanel';
import { environmentStore } from '@/react/hooks/current-environment-store';

import { PageHeader } from '@@/PageHeader';

import { useEnvironment } from '../environments/queries';

import { EnvironmentList } from './EnvironmentList';
import { EdgeLoadingSpinner } from './EdgeLoadingSpinner';
import { MotdPanel } from './MotdPanel';
import { LicenseNodePanel } from './LicenseNodePanel';
import { BackupFailedPanel } from './BackupFailedPanel';

export function HomeView() {
  const { environmentId: currentEnvironmentId } = useStore(environmentStore);
  const environmentQuery = useEnvironment(currentEnvironmentId);

  const [connectingToEdgeEndpoint, setConnectingToEdgeEndpoint] =
    useState(false);

  const router = useRouter();
  const environment = environmentQuery.data;
  return (
    <>
      <PageHeader reload title="Home" breadcrumbs={[{ label: 'Environments' }]}>
        {environment && <DockerSnapshotPanel environment={environment} />}
      </PageHeader>

      {process.env.PORTAINER_EDITION !== 'CE' && <LicenseNodePanel />}

      <MotdPanel />

      {process.env.PORTAINER_EDITION !== 'CE' && <BackupFailedPanel />}

      {connectingToEdgeEndpoint ? (
        <EdgeLoadingSpinner />
      ) : (
        <EnvironmentList
          onClickBrowse={handleBrowseClick}
          onRefresh={confirmTriggerSnapshot}
        />
      )}
    </>
  );

  async function confirmTriggerSnapshot() {
    const result = await confirmEndpointSnapshot();
    if (!result) {
      return;
    }
    try {
      await snapshotEndpoints();
      notifications.success('Success', 'Environments updated');
      router.stateService.reload();
    } catch (err) {
      notifications.error(
        'Failure',
        err as Error,
        'An error occurred during environment snapshot'
      );
    }
  }

  function handleBrowseClick(environment: Environment) {
    if (isEdgeEnvironment(environment.Type)) {
      setConnectingToEdgeEndpoint(true);
    }
  }
}

async function confirmEndpointSnapshot() {
  return confirmAsync({
    title: buildTitle('Are you sure?'),
    message:
      'Triggering a manual refresh will poll each environment to retrieve its information, this may take a few moments.',
    buttons: {
      confirm: {
        label: 'Continue',
        className: 'btn-primary',
      },
    },
  });
}
