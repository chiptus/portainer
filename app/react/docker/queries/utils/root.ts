import { EnvironmentId } from '@/portainer/environments/types';

export const queryKeys = {
  root: (environmentId: EnvironmentId) => ['docker', environmentId] as const,
  snapshot: (environmentId: EnvironmentId) =>
    [...queryKeys.root(environmentId), 'snapshot'] as const,
  snapshotQuery: (environmentId: EnvironmentId) =>
    [...queryKeys.snapshot(environmentId)] as const,
};

export function buildDockerUrl(environmentId: EnvironmentId) {
  return `/docker/${environmentId}`;
}

export function buildDockerSnapshotUrl(environmentId: EnvironmentId) {
  return `${buildDockerUrl(environmentId)}/snapshot`;
}
