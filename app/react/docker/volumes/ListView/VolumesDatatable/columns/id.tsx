import { CellProps, Column } from 'react-table';
import { useSref } from '@uirouter/react';

import { DockerVolume } from '@/react/docker/volumes/types';
import { truncate } from '@/portainer/filters/filters';

export const id: Column<DockerVolume> = {
  Header: 'Id',
  accessor: 'Id',
  Cell,
  id: 'id',
  disableFilters: true,
  canHide: true,
  sortType: 'string',
  Filter: () => null,
};

function Cell({
  value: name,
  row: { original: volume },
}: CellProps<DockerVolume>) {
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
