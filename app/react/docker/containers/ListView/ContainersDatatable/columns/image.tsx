import { Column } from 'react-table';
import { useSref } from '@uirouter/react';

import { ImageStatus } from '@/react/docker/components/ImageStatus';
import type { DockerContainer } from '@/react/docker/containers/types';
import { isOfflineEndpoint } from '@/portainer/helpers/endpointHelper';
import { trimSHA } from '@/docker/filters/utils';

import { useRowContext } from '../RowContext';

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
}

function ImageCell({ value: imageName }: Props) {
  const linkProps = useSref('docker.images.image', { id: imageName });
  const shortImageName = trimSHA(imageName);

  const { environment } = useRowContext();

  if (isOfflineEndpoint(environment)) {
    return <span>{shortImageName}</span>;
  }

  return (
    <a href={linkProps.href} onClick={linkProps.onClick}>
      <ImageStatus imageName={imageName} environmentId={environment.Id} />
      {shortImageName}
    </a>
  );
}
