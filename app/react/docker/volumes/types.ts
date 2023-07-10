import { Volume } from 'docker-types/generated/1.41';

type DecoratedDockerVolume = {
  Id: string;
  StackName: string;
  Used: boolean;
};

export type DockerVolume = DecoratedDockerVolume & Omit<Volume, 'Name'>;
