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
      <span className="blocklist-item-desc space-x-2 vertical-center">
        <Stat
          value={addPlural(snapshot.JobCount, 'job')}
          icon="list"
          featherIcon
        />
        <Stat
          value={addPlural(snapshot.GroupCount, 'group')}
          icon="svg-objectgroup"
        />
        <Stat
          value={addPlural(snapshot.TaskCount, 'task')}
          icon="box"
          featherIcon
        >
          {snapshot.TaskCount > 0 && (
            <span className="space-x-2">
              <Stat
                value={snapshot.RunningTaskCount}
                icon="power"
                featherIcon
                iconClass="icon-success"
              />
              <Stat
                value={snapshot.TaskCount - snapshot.RunningTaskCount}
                icon="power"
                featherIcon
                iconClass="icon-danger"
              />
            </span>
          )}
        </Stat>
      </span>

      <span className="small text-muted space-x-2 vertical-center">
        <span>Nomad</span>
        <Stat
          value={addPlural(snapshot.NodeCount, 'node')}
          icon="hard-drive"
          featherIcon
        />
      </span>
    </div>
  );
}
