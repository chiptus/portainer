import clsx from 'clsx';
import { useQuery } from 'react-query';
import { getStackImagesStatus } from 'Portainer/services/api/stack.service';
import { statusClass } from 'Docker/components/ImageStatus/helpers';

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
