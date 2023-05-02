import { CellContext } from '@tanstack/react-table';

import { GitCredential } from '@/react/portainer/account/git-credentials/types';

import { Link } from '@@/Link';

import { columnHelper } from './helper';

export const name = columnHelper.accessor('name', { cell: NameCell });

export function NameCell({
  getValue,
  row,
}: CellContext<GitCredential, string>) {
  const name = getValue();

  return (
    <Link
      to="portainer.account.gitEditGitCredential"
      params={{ id: row.id }}
      title={name}
    >
      {name}
    </Link>
  );
}
