import { DockerContainer } from '@/react/docker/containers/types';
import { EnvironmentId } from '@/portainer/environments/types';
import { EdgeStack } from '@/react/edge/edge-stacks/types';

export interface ContainersQueryParams {
  edgeStackId?: EdgeStack['Id'];
}

export const queryKeys = {
  root: (environmentId: EnvironmentId) => ['docker', environmentId] as const,
  snapshot: (environmentId: EnvironmentId) =>
    [...queryKeys.root(environmentId), 'snapshot'] as const,
  containers: (environmentId: EnvironmentId) =>
    [...queryKeys.snapshot(environmentId), 'containers'] as const,
  containersQuery: (
    environmentId: EnvironmentId,
    params: ContainersQueryParams
  ) => [...queryKeys.containers(environmentId), params] as const,
  container: (
    environmentId: EnvironmentId,
    containerId: DockerContainer['Id']
  ) => [...queryKeys.containers(environmentId), containerId] as const,
};

export function buildDockerUrl(environmentId: EnvironmentId) {
  return `/docker/${environmentId}`;
}

export function buildDockerSnapshotUrl(environmentId: EnvironmentId) {
  return `${buildDockerUrl(environmentId)}/snapshot`;
}

export function buildDockerSnapshotContainersUrl(
  environmentId: EnvironmentId,
  containerId?: DockerContainer['Id']
) {
  let url = `${buildDockerSnapshotUrl(environmentId)}/containers`;

  if (containerId) {
    url += `/${containerId}`;
  }

  return url;
}
