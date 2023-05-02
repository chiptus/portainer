import { CellContext } from '@tanstack/react-table';
import { useSref } from '@uirouter/react';

import { DockerImage } from '@/react/docker/images/types';
import { truncate } from '@/portainer/filters/filters';

import { columnHelper } from './helper';

export const id = columnHelper.accessor('Id', {
  id: 'id',
  header: 'Id',
  cell: Cell,
});

function Cell({
  getValue,
  row: { original: image },
}: CellContext<DockerImage, string>) {
  const name = getValue();

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
