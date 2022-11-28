import clsx from 'clsx';
import { useQuery } from 'react-query';
import { Loader } from 'lucide-react';

import {
  getContainerImagesStatus,
  getServiceImagesStatus,
} from '@/react/docker/images/image.service';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { statusClass } from '@/react/docker/components/ImageStatus/helpers';
import { ResourceID, ResourceType } from '@/react/docker/images/types';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { Icon } from '@@/Icon';

export interface Props {
  environmentId: EnvironmentId;
  resourceId: ResourceID;
  resourceType?: ResourceType;
  nodeName?: string;
}

export function ImageStatus({
  environmentId,
  resourceId,
  resourceType = ResourceType.CONTAINER,
  nodeName = '',
}: Props) {
  const enableImageNotificationQuery = useEnvironment(
    environmentId,
    (environment) => environment?.EnableImageNotification
  );

  const { data, isLoading, isError } = useImageNotification(
    environmentId,
    resourceId,
    resourceType,
    nodeName,
    enableImageNotificationQuery.data
  );

  if (!enableImageNotificationQuery.data || isError) {
    return null;
  }

  if (isLoading || !data) {
    return <Icon icon={Loader} className="spin !mr-0.5" />;
  }

  return <span className={clsx(statusClass(data), 'space-right')} />;
}

export function useImageNotification(
  environmentId: number,
  resourceId: ResourceID,
  resourceType: ResourceType,
  nodeName: string,
  enabled = false
) {
  return useQuery(
    [
      'environments',
      environmentId,
      'docker',
      'images',
      resourceType,
      resourceId,
      'status',
    ],
    () =>
      resourceType === ResourceType.SERVICE
        ? getServiceImagesStatus(environmentId, resourceId)
        : getContainerImagesStatus(environmentId, resourceId, nodeName),
    {
      enabled,
    }
  );
}
