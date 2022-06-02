import clsx from 'clsx';
import { PropsWithChildren } from 'react';

type Severity = 'info' | 'warning';

interface Props {
  showIcon?: boolean;
  type?: Severity;
  containerStyles?: Record<string, string>;
  iconStyles?: Record<string, string>;
}

const iconClasses: Record<Severity, string> = {
  warning: 'fa-exclamation-triangle orange-icon',
  info: 'fa-info-circle blue-icon',
};

const textClasses: Record<Severity, string> = {
  warning: 'small text-warning',
  info: 'small text-muted',
};

export function Alert({
  children,
  showIcon = true,
  type = 'warning',
  containerStyles = {},
  iconStyles = {},
}: PropsWithChildren<Props>) {
  return (
    <div className={clsx(textClasses[type], 'my-2')} style={containerStyles}>
      {showIcon && (
        <i
          className={clsx(iconClasses[type], 'fa space-right')}
          aria-hidden="true"
          style={iconStyles}
        />
      )}
      {children}
    </div>
  );
}
