import { DockerVolumeResponse } from './types/response';

type DecoratedDockerVolume = {
  Id: string;
  StackName: string;
  Used: boolean;
};

export type DockerVolume = DecoratedDockerVolume &
  Omit<DockerVolumeResponse, 'Name'>;
