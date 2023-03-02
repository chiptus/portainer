import { Key } from 'lucide-react';
import { useStore } from 'zustand';

import { Datatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useSearchBarState } from '@@/datatables/SearchBar';

import { useCloudCredentials } from '../../cloudSettings.service';

import { CredentialsDatatableActions } from './CredentialsDatatableActions';
import { columns } from './columns';

const storageKey = 'cloudCredentials';

const settingsStore = createPersistedStore(storageKey, 'name');

export function CredentialsDatatable() {
  const settings = useStore(settingsStore);
  const [search, setSearch] = useSearchBarState(storageKey);
  const cloudCredentialsQuery = useCloudCredentials();

  const credentials = cloudCredentialsQuery.data || [];

  return (
    <Datatable
      titleIcon={Key}
      title="Shared credentials"
      initialPageSize={settings.pageSize}
      onPageSizeChange={settings.setPageSize}
      initialSortBy={settings.sortBy}
      onSortByChange={settings.setSortBy}
      searchValue={search}
      onSearchChange={setSearch}
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
