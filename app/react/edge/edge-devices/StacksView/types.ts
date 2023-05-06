import { DockerContainer } from '@/react/docker/containers/types';
import { Stack } from '@/react/docker/stacks/types';

export interface StackMetadata {
  isEdgeStack?: boolean;
  isStack?: boolean;
  stackId?: Stack['Id'];
  isExternalStack?: boolean;
}

export type StackInAsyncSnapshot = DockerContainer & {
  // StackMetadata is to determine whether a container
  // is associated with either a Stack or Edge stack
  Metadata: StackMetadata;
};
