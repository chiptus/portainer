import { useCallback, useEffect } from 'react';

import { useSnapshotMutation } from '@/portainer/environments/queries';

import { useCurrentEnvironment } from './useCurrentEnvironment';

export function useCurrentEnvironmentSnapshot() {
  const environmentQuery = useCurrentEnvironment();

  const environment = environmentQuery.data;

  const snapshot = environment?.Nomad?.Snapshots?.length
    ? environment.Nomad.Snapshots[0]
    : null;

  const snapshotMutation = useSnapshotMutation();

  const { mutate } = snapshotMutation;

  const triggerSnapshot = useCallback(() => {
    const envId = environment?.Id;

    if (!envId) {
      return;
    }

    mutate(envId);
  }, [environment?.Id, mutate]);

  useEffect(() => {
    if (environment?.Id) {
      triggerSnapshot();
    }
  }, [environment?.Id, triggerSnapshot]);

  return {
    dashboardQuery: snapshot,
    triggerSnapshot,
    isLoading: snapshotMutation.isLoading || environmentQuery.isLoading,
  };
}
