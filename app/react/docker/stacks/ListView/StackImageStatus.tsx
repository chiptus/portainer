import clsx from 'clsx';
import { useQuery } from 'react-query';

import { getStackImagesStatus } from '@/portainer/services/api/stack.service';
import { statusClass } from '@/docker/components/ImageStatus/helpers';

export interface Props {
  stackId: number;
}

export function StackImageStatus({ stackId }: Props) {
  const { data, isLoading } = useQuery(
    ['stacks', stackId, 'images', 'status'],
    () => getStackImagesStatus(stackId)
  );
  return <span className={clsx(statusClass(data, isLoading), 'space-right')} />;
}
