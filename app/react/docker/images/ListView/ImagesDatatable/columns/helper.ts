import { createColumnHelper } from '@tanstack/react-table';

import { DockerImage } from '@/react/docker/images/types';

export const columnHelper = createColumnHelper<DockerImage>();
