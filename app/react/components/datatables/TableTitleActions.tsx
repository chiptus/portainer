import { Children, PropsWithChildren } from 'react';

import { useTableContext } from './TableContainer';

export function TableTitleActions({ children }: PropsWithChildren<unknown>) {
  useTableContext();

  if (Children.count(children) === 0) {
    return null;
  }

  return <div className="settings">{children}</div>;
}
