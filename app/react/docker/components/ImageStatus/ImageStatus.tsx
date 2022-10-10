import clsx from 'clsx';
import { useQuery } from 'react-query';

import { useEnvironment } from '@/portainer/environments/queries';
import { getImagesStatus } from '@/react/docker/images/image.service';
import { statusClass } from '@/react/docker/components/ImageStatus/helpers';
import { EnvironmentId } from '@/portainer/environments/types';

export interface Props {
  imageName: string;
  environmentId: EnvironmentId;
}

export function ImageStatus({ imageName, environmentId }: Props) {
  const enableImageNotificationQuery = useEnvironment(
    environmentId,
    (environment) => environment?.EnableImageNotification
  );

  const { data, isLoading } = useImageNotification(environmentId, imageName);

  return !isLoading && !enableImageNotificationQuery.data ? null : (
    <span className={clsx(statusClass(data, isLoading), 'space-right')} />
  );
}

export function useImageNotification(environmentId: number, imageName: string) {
  const enableImageNotificationQuery = useEnvironment(
    environmentId,
    (environment) => environment?.EnableImageNotification
  );

  return useQuery(
    ['environments', environmentId, 'docker', 'images', imageName, 'status'],
    () => getImagesStatus(environmentId, imageName),
    {
      enabled: enableImageNotificationQuery.data,
    }
  );
}
