import { CellProps, Column } from 'react-table';
import { Clock, FileText } from 'react-feather';

import { Task } from '@/nomad/types';

import { Link } from '@@/Link';

export const actions: Column<Task> = {
  Header: 'Task Actions',
  id: 'actions',
  disableFilters: true,
  canHide: true,
  disableResizing: true,
  width: '5px',
  sortType: 'string',
  Filter: () => null,
  Cell: ActionsCell,
};

export function ActionsCell({ row }: CellProps<Task>) {
  const params = {
    allocationID: row.original.AllocationID,
    taskName: row.original.TaskName,
    namespace: row.original.Namespace,
    jobID: row.original.JobID,
  };

  return (
    <div className="text-center vertical-center">
      {/* events */}
      <Link
        to="nomad.events"
        params={params}
        title="Events"
        className="space-right"
      >
        <Clock className="feather space-right" />
      </Link>

      {/* logs */}
      <Link to="nomad.logs" params={params} title="Logs">
        <FileText className="feather space-right" />
      </Link>
    </div>
  );
}
