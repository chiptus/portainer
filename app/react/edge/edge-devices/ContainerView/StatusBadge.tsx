import clsx from 'clsx';

import { ContainerStatus } from '@/react/docker/containers/types';

export function StatusBadge({
  status,
  state,
}: {
  status: ContainerStatus;
  state: string;
}) {
  const isRunning = [ContainerStatus.Running, ContainerStatus.Healthy].includes(
    status
  );

  return (
    <>
      <i
        className={clsx('fa fa-heartbeat space-right', {
          'green-icon': isRunning,
          'red-icon': !isRunning,
        })}
      />

      {state}
    </>
  );
}
