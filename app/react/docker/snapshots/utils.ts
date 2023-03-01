import { parseViewModel as parseContainer } from '../containers/utils';
import { parseViewModel as parseImage } from '../images/utils';
import { parseViewModel as parseVolume } from '../volumes/utils';

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
    Volumes: response.Volumes.Volumes.map(parseVolume),
    SnapshotTime: response.Info.SystemTime,
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
