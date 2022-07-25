import { EnvironmentId } from '@/portainer/environments/types';

import { useLogsStatus } from '../queries/useLogsStatus';
import { EdgeStack } from '../types';

interface Props {
  environmentId: EnvironmentId;
  edgeStackId: EdgeStack['Id'];
}

export function ActionStatus({ environmentId, edgeStackId }: Props) {
  const logsStatusQuery = useLogsStatus(edgeStackId, environmentId);

  return <>{getStatusText(logsStatusQuery.data)}</>;
}

function getStatusText(status?: 'pending' | 'collected' | 'idle') {
  switch (status) {
    case 'collected':
      return 'Logs available for download';
    case 'pending':
      return 'Logs marked for collection, please wait until the logs are available';
    default:
      return null;
  }
}
