import { credentialTitles } from '../../../types';

import { columnHelper } from './helper';

export const provider = columnHelper.accessor(
  (row) => credentialTitles[row.provider],
  {
    header: 'Provider',
    id: 'provider',
  }
);
