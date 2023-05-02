import { createColumnHelper } from '@tanstack/react-table';

import { GitCredential } from '../../../git-credentials/types';

export const columnHelper = createColumnHelper<GitCredential>();
