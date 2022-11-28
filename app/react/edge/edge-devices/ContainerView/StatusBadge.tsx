import { HeartPulse } from 'lucide-react';

import { ContainerStatus } from '@/react/docker/containers/types';

import { Icon } from '@@/Icon';

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
      <Icon
        icon={HeartPulse}
        mode={isRunning ? 'success' : 'warning'}
        className="mr-1"
      />
      {state}
    </>
  );
}
