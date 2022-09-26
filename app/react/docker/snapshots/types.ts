import { DockerContainer } from '../containers/types';
import { DockerImage } from '../images/types';
import { DockerVolume } from '../volumes/types';

export type DockerSnapshot = {
  Containers: DockerContainer[];
  Volumes: DockerVolume[];
  Images: DockerImage[];
  SnapshotTime: string;
};
