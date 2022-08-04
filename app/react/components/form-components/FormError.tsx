import { PropsWithChildren } from 'react';

import { Icon } from '@@/Icon';

export function FormError({ children }: PropsWithChildren<unknown>) {
  return (
    <p className="text-muted small vertical-center">
      <Icon icon="alert-triangle" className="icon-warning" feather />
      <span className="text-muted">{children}</span>
    </p>
  );
}
