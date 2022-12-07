import { parseViewModel as parseContainer } from '../containers/utils';
import { parseViewModel as parseImage } from '../images/utils';
import { parseViewModel as parseVolume } from '../volumes/utils';

import { DockerSnapshotRaw } from './types';
import { DockerSnapshotResponse } from './types/response';

export function parseViewModel(
  response: DockerSnapshotResponse
): DockerSnapshotRaw {
  return {
    Containers: response.Containers.map(parseContainer),
    Images: response.Images.map(parseImage),
    Volumes: response.Volumes.Volumes.map(parseVolume),
    SnapshotTime: response.Info.SystemTime,
  };
}
