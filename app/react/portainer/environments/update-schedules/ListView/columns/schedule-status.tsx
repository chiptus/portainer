import { CellContext } from '@tanstack/react-table';

import { Tooltip } from '@@/Tip/Tooltip';

import { EdgeUpdateListItemResponse } from '../../queries/list';
import { StatusType } from '../../types';

import { columnHelper } from './helper';

export const scheduleStatus = columnHelper.accessor('status', {
  header: StatusHeader,
  cell: StatusCell,
});

function StatusHeader() {
  return (
    <>
      Status
      <Tooltip
        position="bottom"
        message={
          <div className="flex flex-col gap-y-2 p-2">
            <span className="font-bold">
              Understanding Edge Agent Update Statuses
            </span>
            <div
              className="grid gap-2"
              style={{ gridTemplateColumns: 'repeat(2, minmax(0, auto))' }}
            >
              <span>1.</span>
              <span>
                <span className="font-bold">Sent: </span>
                The update request has been sent to the Edge Agent but has not
                yet been acknowledged.
              </span>
              <span>2.</span>
              <span>
                <span className="font-bold">Pending: </span>
                The scheduler for the Edge Agent has been successfully
                dispatched to the Edge devices, and they have acknowledged
                receipt. The devices will initiate the update at the scheduled
                time specified in the scheduler.
              </span>
              <span>3.</span>
              <span>
                <span className="font-bold">Updating: </span>
                The Edge Agent is currently undergoing the update process.
                Changes are being applied to bring it to the latest version.
              </span>
              <span>4.</span>
              <span>
                <span className="font-bold">Success: </span>
                The update for the Edge Agent has been successfully completed,
                and it is now running the latest version.
              </span>
            </div>
          </div>
        }
      />
    </>
  );
}

function StatusCell({
  getValue,
  row: {
    original: { statusMessage },
  },
}: CellContext<
  EdgeUpdateListItemResponse,
  EdgeUpdateListItemResponse['status']
>) {
  const status = getValue();

  switch (status) {
    case StatusType.Failed:
      return statusMessage;
    default:
      return StatusType[status];
  }
}
