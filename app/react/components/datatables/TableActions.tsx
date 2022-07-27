import clsx from 'clsx';
import { Children, PropsWithChildren } from 'react';

import { useTableContext } from './TableContainer';

interface Props {
  className?: string;
}

export function TableActions({
  children,
  className,
}: PropsWithChildren<Props>) {
  useTableContext();

  if (Children.count(children) === 0) {
    return null;
  }

  return <div className={clsx('actionBar', className)}>{children}</div>;
}
