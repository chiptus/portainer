import { useSref } from '@uirouter/react';
import { CellContext } from '@tanstack/react-table';

import { DockerVolume } from '@/react/docker/volumes/types';
import { truncate } from '@/portainer/filters/filters';

import { columnHelper } from './helper';

export const id = columnHelper.accessor('Id', {
  header: 'Id',
  id: 'id',
  cell: Cell,
});

function Cell({
  getValue,
  row: { original: volume },
}: CellContext<DockerVolume, string>) {
  const name = getValue();

  const linkProps = useSref('.volume', {
    id: volume.Id,
    volumeId: volume.Id,
  });

  return (
    <>
      <a href={linkProps.href} onClick={linkProps.onClick} title={name}>
        {truncate(name, 40)}
      </a>
      {!volume.Used && (
        <span className="label label-warning image-tag ml-2">Unused</span>
      )}
    </>
  );
}
