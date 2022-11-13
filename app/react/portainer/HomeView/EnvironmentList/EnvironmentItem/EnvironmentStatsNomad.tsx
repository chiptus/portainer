import {
  EnvironmentType,
  NomadSnapshot,
} from '@/react/portainer/environments/types';
import { addPlural } from '@/portainer/helpers/strings';

import { EnvironmentStatsItem } from '../../../../components/EnvironmentStatsItem';

import { AgentVersionTag } from './AgentVersionTag';

interface Props {
  snapshots: NomadSnapshot[];
  type: EnvironmentType;
  agentVersion: string;
}

export function EnvironmentStatsNomad({
  snapshots = [],
  agentVersion,
  type,
}: Props) {
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
        <EnvironmentStatsItem
          value={addPlural(snapshot.JobCount, 'job')}
          icon="list"
          featherIcon
        />
        <EnvironmentStatsItem
          value={addPlural(snapshot.GroupCount, 'group')}
          icon="svg-objectgroup"
        />
        <EnvironmentStatsItem
          value={addPlural(snapshot.TaskCount, 'task')}
          icon="box"
          featherIcon
        >
          {snapshot.TaskCount > 0 && (
            <span className="space-x-2">
              <EnvironmentStatsItem
                value={snapshot.RunningTaskCount}
                icon="power"
                featherIcon
                iconClass="icon-success"
              />
              <EnvironmentStatsItem
                value={snapshot.TaskCount - snapshot.RunningTaskCount}
                icon="power"
                featherIcon
                iconClass="icon-danger"
              />
            </span>
          )}
        </EnvironmentStatsItem>
      </span>

      <span className="small text-muted space-x-2 vertical-center">
        <span>Nomad</span>
        <EnvironmentStatsItem
          value={addPlural(snapshot.NodeCount, 'node')}
          icon="hard-drive"
          featherIcon
        />
        <AgentVersionTag type={type} version={agentVersion} />
      </span>
    </div>
  );
}
