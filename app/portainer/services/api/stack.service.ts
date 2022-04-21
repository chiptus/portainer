import axios from 'Portainer/services/axios';
import { ImageStatus } from 'Docker/components/ImageStatus/types';

export async function getStackImagesStatus(id: number) {
  try {
    const { data } = await axios.get<ImageStatus>(
      `/stacks/${id}/images_status`
    );
    return data;
  } catch (e) {
    return {
      Status: 'unknown',
      Message: `Unable to retrieve image status for stack: ${id}`,
    };
  }
}
