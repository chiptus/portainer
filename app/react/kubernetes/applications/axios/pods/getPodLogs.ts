import _ from 'lodash';

import axios from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { urlBuilder } from '@/react/kubernetes/applications/axios/pods/urlBuilder';
import { parseKubernetesAxiosError } from '@/react/kubernetes/axiosError';

export type PodLogsParams = {
  timestamps?: boolean;
  sinceSeconds?: number;
  tailLines?: number;
  container?: string;
};

export async function getPodLogs(
  environmentId: EnvironmentId,
  namespace: string,
  podId: string,
  params?: PodLogsParams
): Promise<string> {
  try {
    const { data } = await axios.get<string>(
      urlBuilder(environmentId, namespace, podId, 'log'),
      {
        params: _.pickBy(params),
      }
    );

    return data;
  } catch (e) {
    throw parseKubernetesAxiosError(e, 'Unable to get pod logs');
  }
}
