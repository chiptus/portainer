import clsx from 'clsx';
import { useQuery } from 'react-query';
import { Loader } from 'lucide-react';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { getStackImagesStatus } from '@/portainer/services/api/stack.service';
import { statusClass } from '@/react/docker/components/ImageStatus/helpers';

import { Icon } from '@@/Icon';

export interface Props {
  stackId: number;
  environmentId: number;
}

export function StackImageStatus({ stackId, environmentId }: Props) {
  const { data, isLoading, isError } = useStackImageNotification(
    stackId,
    environmentId
  );

  if (isError) {
    return null;
  }

  if (isLoading || !data) {
    return <Icon icon={Loader} className="spin !mr-0.5" />;
  }

  return <span className={clsx(statusClass(data), 'space-right')} />;
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
