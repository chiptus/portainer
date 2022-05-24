import { CellProps, Column } from 'react-table';

import { Environment } from '@/portainer/environments/types';
import { EdgeIndicator } from '@/portainer/home/EnvironmentList/EnvironmentItem/EdgeIndicator';

export const heartbeat: Column<Environment> = {
  Header: 'Heartbeat',
  accessor: 'Status',
  id: 'status',
  Cell: StatusCell,
  disableFilters: true,
  canHide: true,
};

export function StatusCell({
  row: { original: environment },
}: CellProps<Environment>) {
  let checkIn = environment.EdgeCheckinInterval;

  if (environment.Edge.AsyncMode) {
    checkIn = 60;

    const intervals = [
      environment.Edge.PingInterval,
      environment.Edge.SnapshotInterval,
      environment.Edge.CommandInterval,
    ].filter((n) => n > 0);

    if (intervals.length > 0) {
      checkIn = Math.min(...intervals);
    }
  }

  return (
    <EdgeIndicator
      checkInInterval={checkIn}
      edgeId={environment.EdgeID}
      lastCheckInDate={environment.LastCheckInDate}
      queryDate={environment.QueryDate}
    />
  );
}
