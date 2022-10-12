import { formatDuration, intervalToDuration } from 'date-fns';

interface Props {
  start: Date;
  end: Date;
  pattern: string;
  snapshotDisabled?: boolean;
  muted?: boolean;
}
export function IntervalColumn({
  start,
  end,
  pattern,
  snapshotDisabled,
  muted,
}: Props) {
  let intervalStr = '(snapshot is disabled)';

  if (!snapshotDisabled) {
    const interval = formatDuration(intervalToDuration({ start, end }));
    intervalStr = interval === '' ? '(now)' : pattern.replace('$', interval);
  }

  return <span className={muted ? 'text-muted' : ''}>{intervalStr}</span>;
}
