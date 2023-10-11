import _ from 'lodash';

import {
  COMPOSE_STACK_NAME_LABEL,
  SWARM_STACK_NAME_LABEL,
} from '@/react/constants';
import { VolumeViewModel } from '@/docker/models/volume';

import { parseListViewModel as parseContainer } from '../containers/utils';
import { parseViewModel as parseImage } from '../images/utils';

import { DockerContainerSnapshot, DockerSnapshotRaw } from './types';
import {
  DockerContainerSnapshotResponse,
  DockerSnapshotResponse,
} from './types/response';

export function parseViewModel(
  response: DockerSnapshotResponse
): DockerSnapshotRaw {
  return {
    Containers: response.Containers.map(parseDockerContainerSnapshot),
    Images: response.Images.map(parseImage),
    Volumes: response.Volumes.Volumes.map((v) => new VolumeViewModel(v)),
    Info: response.Info,
  };
}

export function parseDockerContainerSnapshot(
  response: DockerContainerSnapshotResponse
): DockerContainerSnapshot {
  return {
    ...parseContainer(response),
    Env: response.Env ? response.Env : [],
  };
}

export function filterUniqueContainersBasedOnStack(
  containers: DockerContainerSnapshot[]
): DockerContainerSnapshot[] {
  return _.uniqBy(
    containers.filter(
      (item) =>
        item.Labels &&
        (item.Labels[COMPOSE_STACK_NAME_LABEL] ||
          item.Labels[SWARM_STACK_NAME_LABEL])
    ),
    'StackName'
  );
}
