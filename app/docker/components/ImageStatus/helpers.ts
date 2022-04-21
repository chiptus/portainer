import { ImageStatus as ImgStatus } from 'Docker/components/ImageStatus/types';
import style from 'Docker/components/ImageStatus/ImageStatus.module.css';

export function statusClass(
  status?: ImgStatus | null,
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
