import { ImageStatus } from '@/react/docker/images/types';

import style from './ImageStatus.module.css';

export function statusClass(
  status?: ImageStatus | null,
  isLoading?: boolean
): string {
  if (isLoading || !status) {
    return 'fa fa-spinner fa-spin';
  }
  switch (status.Status) {
    case 'outdated':
      return style.outdated;
    case 'updated':
      return style.updated;
    case 'processing':
      return style.processing;
    default:
      return style.unknown;
  }
}
