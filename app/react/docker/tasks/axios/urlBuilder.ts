import { TaskId } from '@/react/docker/tasks/types';
import { EnvironmentId } from '@/react/portainer/environments/types';

export function urlBuilder(
  endpointId: EnvironmentId,
  id?: TaskId,
  action?: string
) {
  let url = `/endpoints/${endpointId}/docker/tasks`;

  if (id) {
    url += `/${id}`;
  }

  if (action) {
    url += `/${action}`;
  }

  return url;
}
