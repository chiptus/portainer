import { EnvironmentId } from '@/react/portainer/environments/types';
import axios from '@/portainer/services/axios';

interface ImageRemoveOptions {
  Force: boolean;
  PruneChildren: boolean;
}

interface ImageCommandCreateRequest {
  ImageName: string;
  ImageOperation: string;
  ImageRemoveOptions?: ImageRemoveOptions;
}

export async function removeImage(
  endpointId: EnvironmentId,
  imageId: string,
  force?: boolean,
  pruneChildren?: boolean
) {
  const payload: ImageCommandCreateRequest = {
    ImageName: imageId,
    ImageOperation: 'delete',
    ImageRemoveOptions: {
      Force: force ?? false,
      PruneChildren: pruneChildren ?? false,
    },
  };
  await axios.post<void>(urlBuilder(endpointId), payload);
}

export function urlBuilder(endpointId: EnvironmentId) {
  return `/endpoints/${endpointId}/edge/async/commands/image`;
}
