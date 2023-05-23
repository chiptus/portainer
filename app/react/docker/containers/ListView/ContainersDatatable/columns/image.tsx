import { CellContext } from '@tanstack/react-table';
import { useSref } from '@uirouter/react';

import { ImageStatus } from '@/react/docker/components/ImageStatus';
import type { DockerContainer } from '@/react/docker/containers/types';
import { trimSHA } from '@/docker/filters/utils';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { ImageUpToDateTooltip } from '@/react/docker/components/datatables/TableColumnHeaderImageUpToDate';

import { columnHelper } from './helper';

export const image = columnHelper.accessor('Image', {
  header: ImageHeader,
  id: 'image',
  cell: ImageCell,
});

function ImageHeader() {
  return (
    <>
      Image
      <ImageUpToDateTooltip />
    </>
  );
}

function ImageCell({ getValue, row }: CellContext<DockerContainer, string>) {
  const imageName = getValue();
  const linkProps = useSref('docker.images.image', { id: imageName });
  const shortImageName = trimSHA(imageName);
  const environmentId = useEnvironmentId();

  return (
    <a href={linkProps.href} onClick={linkProps.onClick}>
      <ImageStatus
        environmentId={environmentId}
        resourceId={row.original.Id}
        nodeName={row.original.NodeName}
      />
      {shortImageName}
    </a>
  );
}
