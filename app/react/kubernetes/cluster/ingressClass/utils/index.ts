import { EnvironmentId } from '@/portainer/environments/types';
import PortainerError from '@/portainer/error';
import axios from '@/portainer/services/axios';

import { IngressControllerClassMap } from '../types';

// get all supported ingress classes and controllers, then match them to create a map of ingress classes to controllers
export async function getIngressControllerClassMap(
  environmentId: EnvironmentId
) {
  try {
    const { data: controllerMaps } = await axios.get<
      IngressControllerClassMap[]
    >(`kubernetes/${environmentId}/ingresscontrollers`);
    return controllerMaps;
  } catch (e) {
    throw new PortainerError('Unable to retrieve pods', e as Error);
  }
}
