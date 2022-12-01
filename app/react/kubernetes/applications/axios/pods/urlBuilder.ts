import { EnvironmentId } from '@/react/portainer/environments/types';

export function urlBuilder(
  endpointId: EnvironmentId,
  namespace: string,
  id?: string,
  action?: string
) {
  const namespaceURL = namespace ? `/namespaces/${namespace}` : '';
  let url = `/endpoints/${endpointId}/kubernetes/api/v1${namespaceURL}/pods`;

  if (id) {
    url += `/${id}`;
  }

  if (action) {
    url += `/${action}`;
  }

  return url;
}
