import { CellContext } from '@tanstack/react-table';
import { AlertCircle } from 'lucide-react';

import { Icon } from '@@/Icon';
import { TooltipWithChildren } from '@@/Tip/TooltipWithChildren';

import { StackInAsyncSnapshot } from '../types';

import { columnHelper } from './helper';

export const control = columnHelper.accessor(() => '', {
  header: 'Control',
  id: 'control',
  cell: ControlCell,
});

function ControlCell({
  row: { original: container },
}: CellContext<StackInAsyncSnapshot, string>) {
  if (
    container.Metadata.isEdgeStack ||
    container.Labels['io.portainer.agent'] ||
    container.Metadata.isExternalStack
  ) {
    return (
      <TooltipWithChildren message="This stack was created outside of Portainer. Control over this stack is limited.">
        <span className="vertical-center">
          Limited
          <Icon icon={AlertCircle} mode="warning" />
        </span>
      </TooltipWithChildren>
    );
  }

  return 'Total';
}
