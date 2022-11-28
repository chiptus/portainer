import { Power } from 'lucide-react';

import { EnvironmentStatsItem } from '@@/EnvironmentStatsItem';

interface Props {
  running: number;
  stopped: number;
}

export function RunningStatus({ running, stopped }: Props) {
  return (
    <div>
      <div>
        <EnvironmentStatsItem
          value={`${running || '-'} running`}
          icon={Power}
          iconClass="icon-success"
        />
      </div>
      <div>
        <EnvironmentStatsItem
          value={`${stopped || '-'} stopped`}
          icon={Power}
          iconClass="icon-danger"
        />
      </div>
    </div>
  );
}
