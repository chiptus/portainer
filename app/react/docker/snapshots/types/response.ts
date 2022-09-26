import { DockerContainerResponse } from '@/react/docker/containers/types/response';
import { DockerImageResponse } from '@/react/docker/images/types/response';
import { DockerVolumeResponse } from '@/react/docker/volumes/types/response';

export type DockerSnapshotResponse = {
  Containers: DockerContainerResponse[];
  Images: DockerImageResponse[];
  Info: { SystemTime: string };
  Volumes: { Volumes: DockerVolumeResponse[] };
};
