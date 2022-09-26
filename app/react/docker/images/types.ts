import { DockerImageResponse } from './types/response';

type Status = 'outdated' | 'updated' | 'inprocess' | string;

export interface ImageStatus {
  Status: Status;
  Message: string;
}

type DecoratedDockerImage = {
  Used: boolean;
};

export type DockerImage = DecoratedDockerImage &
  Omit<DockerImageResponse, keyof DecoratedDockerImage>;
