import ReactTooltip from 'react-tooltip';
import clsx from 'clsx';
import _ from 'lodash';
import ReactDOMServer from 'react-dom/server';

import styles from './TooltipWithChildren.module.css';

type Position = 'top' | 'right' | 'bottom' | 'left';

export interface Props {
  position?: Position;
  message: string;
  className?: string;
  children: React.ReactNode;
  heading?: string;
  wrapperClassName?: string;
}

export function TooltipWithChildren({
  message,
  position = 'bottom',
  className = '',
  wrapperClassName = '',
  children,
  heading = '',
}: Props) {
  const id = _.uniqueId('tooltip-');

  const messageHTML = (
    <div className={styles.tooltipContainer}>
      {heading && (
        <div className="w-full mb-3 inline-flex justify-between">
          <span>{heading}</span>
        </div>
      )}
      <div>{message}</div>
    </div>
  );

  return (
    <span
      data-html
      data-multiline
      data-tip={ReactDOMServer.renderToString(messageHTML)}
      data-for={id}
      className={clsx(styles.icon, wrapperClassName, 'inline-flex text-base')}
    >
      {children}
      <ReactTooltip
        id={id}
        multiline
        type="info"
        place={position}
        effect="solid"
        className={clsx(styles.tooltip, className)}
        arrowColor="var(--bg-tooltip-color)"
        delayHide={400}
        clickable
      />
    </span>
  );
}
