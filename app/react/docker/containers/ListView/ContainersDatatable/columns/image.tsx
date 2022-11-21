import { Column, Row } from 'react-table';
import { useSref } from '@uirouter/react';

import { ImageStatus } from '@/react/docker/components/ImageStatus';
import type { DockerContainer } from '@/react/docker/containers/types';
import { trimSHA } from '@/docker/filters/utils';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';

export const image: Column<DockerContainer> = {
  Header: 'Image',
  accessor: 'Image',
  id: 'image',
  disableFilters: true,
  Cell: ImageCell,
  canHide: true,
  sortType: 'string',
  Filter: () => null,
};

interface Props {
  value: string;
  row: Row<DockerContainer>;
}

function ImageCell({ value: imageName, row }: Props) {
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
