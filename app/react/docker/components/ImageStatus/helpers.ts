import { ImageStatus } from '../../images/types';

import style from './ImageStatus.module.css';

export function statusClass(status: ImageStatus) {
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
