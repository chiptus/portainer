import axios from 'Portainer/services/axios';
import { EnvironmentId } from 'Portainer/environments/types';

import { ImageStatus } from './types';

export async function getImagesStatus(
  environmentId: EnvironmentId,
  imageName: string
) {
  try {
    const { data } = await axios.post<ImageStatus>(
      `/docker/${environmentId}/images/status`,
      {
        ImageName: imageName,
      }
    );
    return data;
  } catch (e) {
    return {
      Status: 'unknown',
      Message: `Unable to retrieve image status for image name: ${imageName}`,
    };
  }
}