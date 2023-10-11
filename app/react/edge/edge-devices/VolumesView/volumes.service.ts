import { EnvironmentId } from '@/react/portainer/environments/types';
import axios from '@/portainer/services/axios';

interface VolumeCommandCreateRequest {
  VolumeName: string;
  VolumeOperation: string;
  ForceRemove: boolean;
}

export async function removeVolume(
  endpointId: EnvironmentId,
  volumeId: string
) {
  const payload: VolumeCommandCreateRequest = {
    VolumeName: volumeId,
    VolumeOperation: 'delete',
    ForceRemove: false,
  };
  await axios.post<void>(urlBuilder(endpointId), payload);
}

export function urlBuilder(endpointId: EnvironmentId) {
  return `/endpoints/${endpointId}/edge/async/commands/volume`;
}
