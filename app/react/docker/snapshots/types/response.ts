import { Volume } from 'docker-types/generated/1.41';

import { DockerContainerResponse } from '@/react/docker/containers/types/response';
import { DockerImageResponse } from '@/react/docker/images/types/response';
import { DockerInfoResponse } from '@/react/docker/DashboardView/types/response';

export type DockerContainerSnapshotResponse = DockerContainerResponse & {
  Env?: string[];
};

export type DockerSnapshotResponse = {
  Containers: DockerContainerSnapshotResponse[];
  Images: DockerImageResponse[];
  Info: DockerInfoResponse;
  Volumes: { Volumes: Volume[] };
};
