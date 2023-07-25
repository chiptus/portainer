import { CellContext } from '@tanstack/react-table';

import { ProgressBar } from '@@/ProgressBar';

import { EdgeConfiguration } from '../../types';

import { columnHelper } from './helper';

export const progress = columnHelper.accessor('progress', {
  header: 'Progress',
  cell: ProgressCell,
});

function ProgressCell({
  getValue,
}: CellContext<EdgeConfiguration, EdgeConfiguration['progress']>) {
  const { success, total } = getValue();

  return (
    <>
      <ProgressBar
        className="!w-1/2"
        steps={[
          {
            value: success,
            color: 'var(--ui-success-7)',
          },
          {
            value: total,
            color: 'var(--ui-gray-5)',
          },
        ]}
        total={total}
      />
      <span>
        ( {success} / {total} )
      </span>
    </>
  );
}
