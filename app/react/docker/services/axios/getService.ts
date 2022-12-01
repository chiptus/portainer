import axios from '@/portainer/services/axios';
import PortainerError from '@/portainer/error';
import { urlBuilder } from '@/react/docker/services/axios/urlBuilder';
import {
  DockerServiceResponse,
  ServiceId,
} from '@/react/docker/services/types';
import { EnvironmentId } from '@/react/portainer/environments/types';

export async function getService(
  environmentId: EnvironmentId,
  serviceId: ServiceId
) {
  try {
    const { data } = await axios.get<DockerServiceResponse>(
      urlBuilder(environmentId, serviceId)
    );

    return data;
  } catch (e) {
    throw new PortainerError('Unable to get service', e as Error);
  }
}
