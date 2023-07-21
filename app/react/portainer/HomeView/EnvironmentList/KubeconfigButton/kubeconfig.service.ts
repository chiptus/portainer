import { saveAs } from 'file-saver';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { NodeMetrics, NodeMetric } from '@/react/kubernetes/services/types';

const baseUrl = 'kubernetes';

export async function downloadKubeconfigFile(environmentIds: EnvironmentId[]) {
  try {
    const { headers, data } = await axios.get<Blob>(`${baseUrl}/config`, {
      params: { ids: JSON.stringify(environmentIds) },
      responseType: 'blob',
      headers: {
        Accept: 'text/yaml',
      },
    });
    const contentDispositionHeader = headers['content-disposition'];
    const filename = contentDispositionHeader.replace('attachment;', '').trim();
    saveAs(data, filename);
  } catch (e) {
    throw parseAxiosError(e as Error, '');
  }
}

export async function getMetricsForAllNodes(environmentId: EnvironmentId) {
  try {
    const { data: nodes } = await axios.get<NodeMetrics>(
      `kubernetes/${environmentId}/metrics/nodes`,
      {}
    );
    return nodes;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve services');
  }
}

export async function getMetricsForNode(
  environmentId: EnvironmentId,
  nodeName: string
) {
  try {
    const { data: node } = await axios.get<NodeMetric>(
      `kubernetes/${environmentId}/metrics/nodes/${nodeName}`,
      {}
    );
    return node;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve services');
  }
}

export async function getMetricsForAllPods(
  environmentId: EnvironmentId,
  namespace: string
) {
  try {
    const { data: pods } = await axios.get(
      `kubernetes/${environmentId}/metrics/pods/namespace/${namespace}`,
      {}
    );
    return pods;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve services');
  }
}

export async function getMetricsForPod(
  environmentId: EnvironmentId,
  namespace: string,
  podName: string
) {
  try {
    const { data: pod } = await axios.get(
      `kubernetes/${environmentId}/metrics/pods/namespace/${namespace}/${podName}`,
      {}
    );
    return pod;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve services');
  }
}
