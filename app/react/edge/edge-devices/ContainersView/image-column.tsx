import { ColumnDef } from '@tanstack/react-table';

import { DockerContainer } from '@/react/docker/containers/types';
import { trimSHA } from '@/docker/filters/utils';

export const image: ColumnDef<DockerContainer> = {
  header: 'Image',
  accessorFn: (row) => trimSHA(row.Image),
  id: 'image',
  enableHiding: true,
};
