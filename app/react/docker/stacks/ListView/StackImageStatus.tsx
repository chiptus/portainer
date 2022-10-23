import clsx from 'clsx';
import { useQuery } from 'react-query';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { getStackImagesStatus } from '@/portainer/services/api/stack.service';
import { statusClass } from '@/react/docker/components/ImageStatus/helpers';

export interface Props {
  stackId: number;
  environmentId: number;
}

export function StackImageStatus({ stackId, environmentId }: Props) {
  const { data, isLoading } = useStackImageNotification(stackId, environmentId);

  return <span className={clsx(statusClass(data, isLoading), 'space-right')} />;
}

export function useStackImageNotification(
  stackId: number,
  environmentId?: EnvironmentId
) {
  const enableImageNotificationQuery = useEnvironment(
    environmentId,
    (environment) => environment?.EnableImageNotification
  );

  return useQuery(
    ['stacks', stackId, 'images', 'status'],
    () => getStackImagesStatus(stackId),
    {
      enabled: enableImageNotificationQuery.data,
    }
  );
}
