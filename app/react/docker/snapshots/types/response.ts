import { DockerContainerResponse } from '@/react/docker/containers/types/response';
import { DockerImageResponse } from '@/react/docker/images/types/response';
import { DockerVolumeResponse } from '@/react/docker/volumes/types/response';

export type DockerContainerSnapshotResponse = DockerContainerResponse & {
  Env?: string[];
};

export type DockerSnapshotResponse = {
  Containers: DockerContainerSnapshotResponse[];
  Images: DockerImageResponse[];
  Info: { SystemTime: string };
  Volumes: { Volumes: DockerVolumeResponse[] };
};
