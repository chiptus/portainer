import { EnvironmentId } from '@/portainer/environments/types';
import axios from '@/portainer/services/axios';
import { DockerVolume } from '@/react/docker/volumes/types';

interface VolumeCommandCreateRequest {
  VolumeName: string;
  VolumeOperation: string;
  ForceRemove: boolean;
}

export async function removeVolume(
  endpointId: EnvironmentId,
  volume: DockerVolume
) {
  const payload: VolumeCommandCreateRequest = {
    VolumeName: volume.Id,
    VolumeOperation: 'delete',
    ForceRemove: false,
  };
  await axios.post<void>(urlBuilder(endpointId), payload);
}

export function urlBuilder(endpointId: EnvironmentId) {
  return `/endpoints/${endpointId}/edge/async/commands/volume`;
}
