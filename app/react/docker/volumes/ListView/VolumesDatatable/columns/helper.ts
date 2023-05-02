import { createColumnHelper } from '@tanstack/react-table';

import { DockerVolume } from '@/react/docker/volumes/types';

export const columnHelper = createColumnHelper<DockerVolume>();
