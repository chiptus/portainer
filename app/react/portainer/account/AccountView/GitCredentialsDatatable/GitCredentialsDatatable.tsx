import { Key } from 'lucide-react';

import { useUser } from '@/react/hooks/useUser';
import { useGitCredentials } from '@/react/portainer/account/git-credentials/git-credentials.service';

import { Datatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { columns } from './columns';
import { GitCredentialsDatatableActions } from './GitCredentialsDatatableActions';

const storageKey = 'gitCredentials';

const settingsStore = createPersistedStore(storageKey);

export function GitCredentialsDatatable() {
  const { user } = useUser();
  const gitCredentialsQuery = useGitCredentials(user.Id);

  const tableState = useTableState(settingsStore, storageKey);

  return (
    <Datatable
      dataset={gitCredentialsQuery.data || []}
      settingsManager={tableState}
      columns={columns}
      title="Git credentials"
      titleIcon={Key}
      renderTableActions={(selectedRows) => (
        <GitCredentialsDatatableActions selectedItems={selectedRows} />
      )}
      emptyContentLabel="No credentials found"
      isLoading={gitCredentialsQuery.isLoading}
    />
  );
}
