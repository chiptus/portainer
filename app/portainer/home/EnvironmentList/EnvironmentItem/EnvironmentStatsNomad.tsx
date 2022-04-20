import { NomadSnapshot } from '@/portainer/environments/types';
import { addPlural } from '@/portainer/helpers/strings';

import { Stat } from './EnvironmentStatsItem';

interface Props {
  snapshots: NomadSnapshot[];
}

export function EnvironmentStatsNomad({ snapshots = [] }: Props) {
  if (snapshots.length === 0) {
    return (
      <div className="blocklist-item-line endpoint-item">
        <span className="blocklist-item-desc"> - </span>
      </div>
    );
  }

  const snapshot = snapshots[0];

  return (
    <div className="blocklist-item-line endpoint-item">
      <span className="blocklist-item-desc space-x-4">
        <Stat value={addPlural(snapshot.JobCount, 'job')} icon="fa-th-list" />
        <Stat
          value={addPlural(snapshot.GroupCount, 'group')}
          icon="fa-list-alt"
        />
        <Stat value={addPlural(snapshot.TaskCount, 'task')} icon="fa-cubes">
          {snapshot.TaskCount > 0 && (
            <span className="space-x-2">
              <span>-</span>
              <Stat
                value={snapshot.RunningTaskCount}
                icon="fa-power-off green-icon"
              />
              <Stat
                value={snapshot.TaskCount - snapshot.RunningTaskCount}
                icon="fa-power-off red-icon"
              />
            </span>
          )}
        </Stat>
      </span>

      <span className="small text-muted space-x-3">
        <span>Nomad</span>
        <Stat value={addPlural(snapshot.NodeCount, 'node')} icon="fa-hdd" />
      </span>
    </div>
  );
}
