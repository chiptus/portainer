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
          icon="power"
          featherIcon
          iconClass="icon-success"
        />
      </div>
      <div>
        <EnvironmentStatsItem
          value={`${stopped || '-'} stopped`}
          icon="power"
          featherIcon
          iconClass="icon-danger"
        />
      </div>
    </div>
  );
}
