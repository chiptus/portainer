import { EnvironmentId } from '@/portainer/environments/types';
import axios from '@/portainer/services/axios';
import { ContainerId, DockerContainer } from '@/react/docker/containers/types';

interface ContainerStartOptions {
  CheckpointID: string;
  CheckpointDir: string;
}

interface ContainerRemoveOptions {
  RemoveVolumes: boolean;
  RemoveLinks: boolean;
  Force: boolean;
}

interface CreateContainerCommandRequest {
  ContainerName: string;
  ContainerStartOptions?: ContainerStartOptions;
  ContainerRemoveOptions?: ContainerRemoveOptions;
  ContainerOperation: string;
}

export async function startContainer(
  environmentId: EnvironmentId,
  id: ContainerId
) {
  const payload: CreateContainerCommandRequest = {
    ContainerName: id,
    ContainerOperation: 'start',
  };

  await axios.post<void>(urlBuilder(environmentId), payload);
}

export async function stopContainer(
  endpointId: EnvironmentId,
  id: ContainerId
) {
  const payload: CreateContainerCommandRequest = {
    ContainerName: id,
    ContainerOperation: 'stop',
  };

  await axios.post<void>(urlBuilder(endpointId), payload);
}

export async function restartContainer(
  endpointId: EnvironmentId,
  id: ContainerId
) {
  const payload: CreateContainerCommandRequest = {
    ContainerName: id,
    ContainerOperation: 'restart',
  };

  await axios.post<void>(urlBuilder(endpointId), payload);
}

export async function killContainer(
  endpointId: EnvironmentId,
  id: ContainerId
) {
  const payload: CreateContainerCommandRequest = {
    ContainerName: id,
    ContainerOperation: 'kill',
  };
  await axios.post<void>(urlBuilder(endpointId), payload);
}

export async function removeContainer(
  endpointId: EnvironmentId,
  container: DockerContainer,
  removeVolumes: boolean
) {
  const payload: CreateContainerCommandRequest = {
    ContainerName: container.Id,
    ContainerOperation: 'remove',
    ContainerRemoveOptions: {
      RemoveVolumes: removeVolumes,
      RemoveLinks: false,
      Force: false,
    },
  };
  await axios.post<void>(urlBuilder(endpointId), payload);
}

export function urlBuilder(endpointId: EnvironmentId) {
  return `/endpoints/${endpointId}/edge/async/commands/container`;
}
