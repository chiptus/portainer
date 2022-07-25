import { EnvironmentId } from '@/portainer/environments/types';

import { EdgeStack } from '../types';

import { LogsActions } from './LogsActions';

interface Props {
  asyncMode: boolean;
  environmentId: EnvironmentId;
  edgeStackId: EdgeStack['Id'];
}

export function EnvironmentActions({
  environmentId,
  edgeStackId,
  asyncMode,
}: Props) {
  return (
    <div className="space-x-2">
      {asyncMode && (
        <LogsActions environmentId={environmentId} edgeStackId={edgeStackId} />
      )}
    </div>
  );
}
