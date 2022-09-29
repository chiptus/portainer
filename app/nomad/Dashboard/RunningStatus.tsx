import { Stat } from '@/portainer/home/EnvironmentList/EnvironmentItem/EnvironmentStatsItem';

interface Props {
  running: number;
  stopped: number;
}

export function RunningStatus({ running, stopped }: Props) {
  return (
    <div>
      <div>
        <Stat
          value={`${running || '-'} running`}
          icon="power"
          featherIcon
          iconClass="icon-success"
        />
      </div>
      <div>
        <Stat
          value={`${stopped || '-'} stopped`}
          icon="power"
          featherIcon
          iconClass="icon-danger"
        />
      </div>
    </div>
  );
}
