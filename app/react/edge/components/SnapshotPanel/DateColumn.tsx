import { format } from 'date-fns';

interface Props {
  date: Date;
  snapshotDisabled?: boolean;
}

export function DateColumn({ date, snapshotDisabled }: Props) {
  let formattedDate = '-';

  if (!snapshotDisabled) {
    formattedDate = format(date, 'yyyy/MM/dd HH:mm:ss');
  }

  return <span className="pr-1">{formattedDate}</span>;
}
