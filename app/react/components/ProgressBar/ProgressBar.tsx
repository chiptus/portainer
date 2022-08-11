import clsx from 'clsx';

import styles from './ProgressBar.module.css';

interface Props {
  current: number;
  total: number;
}

export function ProgressBar({ current, total }: Props) {
  const percent = current > total ? 100 : Math.floor((current / total) * 100);

  const inlineStyle =
    current > total
      ? {
          width: '100%',
          backgroundColor: '#f04438',
        }
      : {
          width: `${percent}%`,
          backgroundColor: '#0086c9',
        };

  const progressStyle =
    current > total
      ? clsx('progress', styles.progressAlert)
      : clsx('progress', styles.progressInfo);

  return (
    <div className={progressStyle}>
      <div
        className="progress-bar"
        role="progressbar"
        style={inlineStyle}
        aria-valuenow={percent}
        aria-valuemin={0}
        aria-valuemax={100}
      >
        {' '}
      </div>
    </div>
  );
}
