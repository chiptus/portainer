import clsx from 'clsx';
import { useQuery } from 'react-query';

import { getImagesStatus } from '@/react/docker/images/image.service';
import { statusClass } from '@/react/docker/components/ImageStatus/helpers';
import { EnvironmentId } from '@/portainer/environments/types';

export interface Props {
  imageName: string;
  environmentId: EnvironmentId;
}

export function ImageStatus({ imageName, environmentId }: Props) {
  const { data, isLoading } = useQuery(
    ['environments', environmentId, 'docker', 'images', imageName, 'status'],
    () => getImagesStatus(environmentId, imageName)
  );
  return <span className={clsx(statusClass(data, isLoading), 'space-right')} />;
}
