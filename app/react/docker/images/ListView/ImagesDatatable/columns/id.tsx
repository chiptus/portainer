import { CellProps, Column } from 'react-table';
import { useSref } from '@uirouter/react';

import { DockerImage } from '@/react/docker/images/types';
import { truncate } from '@/portainer/filters/filters';

export const id: Column<DockerImage> = {
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
  row: { original: image },
}: CellProps<DockerImage>) {
  const linkProps = useSref('.image', {
    id: image.Id,
    imageId: image.Id,
  });

  return (
    <>
      <a href={linkProps.href} onClick={linkProps.onClick} title={name}>
        {truncate(name, 40)}
      </a>
      {!image.Used && (
        <span className="label label-warning image-tag ml-2">Unused</span>
      )}
    </>
  );
}
