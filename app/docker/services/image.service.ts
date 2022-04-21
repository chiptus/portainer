import axios from 'Portainer/services/axios';
import { ImageStatus } from 'Docker/components/ImageStatus/types';
import { EnvironmentId } from 'Portainer/environments/types';

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
