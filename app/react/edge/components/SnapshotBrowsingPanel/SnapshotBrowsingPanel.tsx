import clsx from 'clsx';
import { formatDuration, intervalToDuration } from 'date-fns';
import { AlertTriangle } from 'react-feather';

import { isoDate } from '@/portainer/filters/filters';

import styles from './SnapshotBrowsingPanel.module.css';

interface Props {
  snapshotTime: string;
}

export function SnapshotBrowsingPanel({ snapshotTime }: Props) {
  const duration = intervalToDuration({
    start: new Date(snapshotTime),
    end: new Date(),
  });

  const durationStr = formatDuration(duration);

  return (
    <div className={clsx(styles.container)}>
      <div className={clsx(styles.item, 'vertical-center')}>
        <AlertTriangle className="icon icon-sm icon-warning" />
        <span className="text-muted">
          You are browsing a snapshot of the environment taken on{' '}
          {isoDate(snapshotTime)} (
          {durationStr !== '' ? `${durationStr} ago` : 'now'})
        </span>
      </div>
    </div>
  );
}
