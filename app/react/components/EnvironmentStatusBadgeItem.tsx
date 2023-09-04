import clsx from 'clsx';
import { AriaAttributes, PropsWithChildren } from 'react';

import { Icon, IconProps } from '@@/Icon';

export function EnvironmentStatusBadgeItem({
  className,
  children,
  color = 'default',
  icon,
  ...aria
}: PropsWithChildren<
  {
    className?: string;
    color?: 'success' | 'danger' | 'default';
    icon?: IconProps['icon'];
  } & AriaAttributes
>) {
  return (
    <span
      className={clsx(
        'flex items-center gap-1',
        'rounded border-2 border-solid',
        'w-fit px-1 py-px',
        'text-xs font-medium text-gray-7 th-dark:!text-white',
        {
          'border-green-3 bg-green-2 th-dark:border-success-9 th-dark:bg-success-9':
            color === 'success',
          'border-error-3 bg-error-2 th-dark:border-error-9 th-dark:bg-error-9':
            color === 'danger',
        },
        className
      )}
      // eslint-disable-next-line react/jsx-props-no-spreading
      {...aria}
    >
      {icon && (
        <Icon
          icon={icon}
          className={clsx('th-dark:!text-white', {
            '!text-green-7': color === 'success',
            '!text-error-7': color === 'danger',
          })}
        />
      )}

      {children}
    </span>
  );
}
