import clsx from 'clsx';

import styles from './ProgressBar.module.css';

type Step = { value: number; color: string };
interface Props {
  steps: Array<Step>;
  total: number;
}

export function ProgressBar({ steps, total }: Props) {
  const { steps: reducedSteps } = steps.reduce<{
    steps: Array<Step & { percent: number }>;
    total: number;
  }>(
    (acc, cur) => {
      const value =
        acc.total + cur.value > total ? total - acc.total : cur.value;
      return {
        steps: [
          ...acc.steps,
          {
            ...cur,
            value,
            percent: Math.floor((value / total) * 100),
          },
        ],
        total: acc.total + value,
      };
    },
    { steps: [], total: 0 }
  );

  const sum = steps.reduce((sum, s) => sum + s.value, 0);

  return (
    <div
      className={clsx(
        'progress shadow-none',
        sum > 100 ? styles.progressAlert : styles.progressInfo
      )}
      aria-valuemin={0}
      aria-valuemax={100}
      role="progressbar"
    >
      {reducedSteps.map((step, index) => (
        <div
          key={index}
          className="progress-bar shadow-none"
          style={{
            width: `${step.percent}%`,
            backgroundColor: step.color,
          }}
        />
      ))}
    </div>
  );
}
