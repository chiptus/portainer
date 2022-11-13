import {
  DockerSnapshot,
  EnvironmentType,
} from '@/react/portainer/environments/types';
import { addPlural } from '@/portainer/helpers/strings';

import { EnvironmentStatsItem } from '../../../../components/EnvironmentStatsItem';

import { AgentVersionTag } from './AgentVersionTag';

interface Props {
  snapshots: DockerSnapshot[];
  type: EnvironmentType;
  agentVersion: string;
}

export function EnvironmentStatsDocker({
  snapshots = [],
  type,
  agentVersion,
}: Props) {
  if (snapshots.length === 0) {
    return (
      <div className="blocklist-item-line endpoint-item">
        <span className="blocklist-item-desc">No snapshot available</span>
      </div>
    );
  }

  const snapshot = snapshots[0];

  return (
    <div className="blocklist-item-line endpoint-item">
      <span className="blocklist-item-desc">
        <EnvironmentStatsItem
          value={addPlural(snapshot.StackCount, 'stack')}
          icon="layers"
          featherIcon
        />

        {!!snapshot.Swarm && (
          <EnvironmentStatsItem
            value={addPlural(snapshot.ServiceCount, 'service')}
            icon="shuffle"
            featherIcon
          />
        )}

        <ContainerStats
          running={snapshot.RunningContainerCount}
          stopped={snapshot.StoppedContainerCount}
          healthy={snapshot.HealthyContainerCount}
          unhealthy={snapshot.UnhealthyContainerCount}
        />
        <EnvironmentStatsItem
          value={addPlural(snapshot.VolumeCount, 'volume')}
          icon="database"
          featherIcon
        />
        <EnvironmentStatsItem
          value={addPlural(snapshot.ImageCount, 'image')}
          icon="list"
          featherIcon
        />
      </span>

      <span className="small text-muted space-x-3">
        <span>
          {snapshot.Swarm ? 'Swarm' : 'Standalone'} {snapshot.DockerVersion}
        </span>
        {snapshot.Swarm && (
          <EnvironmentStatsItem
            value={addPlural(snapshot.NodeCount, 'node')}
            icon="hard-drive"
            featherIcon
          />
        )}
        <AgentVersionTag version={agentVersion} type={type} />
      </span>
    </div>
  );
}

interface ContainerStatsProps {
  running: number;
  stopped: number;
  healthy: number;
  unhealthy: number;
}

function ContainerStats({
  running,
  stopped,
  healthy,
  unhealthy,
}: ContainerStatsProps) {
  const containersCount = running + stopped;

  return (
    <EnvironmentStatsItem
      value={addPlural(containersCount, 'container')}
      icon="box"
      featherIcon
    >
      {containersCount > 0 && (
        <span className="space-x-2 space-right">
          <EnvironmentStatsItem
            value={running}
            icon="power"
            featherIcon
            iconClass="icon-success"
          />
          <EnvironmentStatsItem
            value={stopped}
            icon="power"
            featherIcon
            iconClass="icon-danger"
          />
          <EnvironmentStatsItem
            value={healthy}
            icon="heart"
            featherIcon
            iconClass="icon-success"
          />
          <EnvironmentStatsItem
            value={unhealthy}
            icon="heart"
            featherIcon
            iconClass="icon-warning"
          />
        </span>
      )}
    </EnvironmentStatsItem>
  );
}
