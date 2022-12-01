import _ from 'lodash';

import axios from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import PortainerError from '@/portainer/error';
import { urlBuilder } from '@/react/docker/services/axios/urlBuilder';
import { ServiceId, ServiceLogsParams } from '@/react/docker/services/types';

export async function getServiceLogs(
  environmentId: EnvironmentId,
  serviceId: ServiceId,
  params?: ServiceLogsParams
): Promise<string> {
  try {
    const { data } = await axios.get<string>(
      urlBuilder(environmentId, serviceId, 'logs'),
      {
        params: _.pickBy(params),
      }
    );

    return data;
  } catch (e) {
    throw new PortainerError('Unable to get service logs', e as Error);
  }
}
