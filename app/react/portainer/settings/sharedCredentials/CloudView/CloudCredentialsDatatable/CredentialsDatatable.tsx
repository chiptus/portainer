import { Key } from 'lucide-react';

import { Datatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { useCloudCredentials } from '../../cloudSettings.service';
import { Credential } from '../../types';

import { CredentialsDatatableActions } from './CredentialsDatatableActions';
import { columns } from './columns';

const storageKey = 'cloudCredentials';

const settingsStore = createPersistedStore(storageKey, 'name');

export function CredentialsDatatable() {
  const tableState = useTableState(settingsStore, storageKey);

  const cloudCredentialsQuery = useCloudCredentials();

  const credentials = cloudCredentialsQuery.data || [];

  return (
    <Datatable<Credential>
      titleIcon={Key}
      title="Shared credentials"
      settingsManager={tableState}
      columns={columns}
      renderTableActions={(selectedRows) => (
        <CredentialsDatatableActions selectedItems={selectedRows} />
      )}
      dataset={credentials}
      emptyContentLabel="No credentials found"
      isLoading={cloudCredentialsQuery.isLoading}
    />
  );
}
