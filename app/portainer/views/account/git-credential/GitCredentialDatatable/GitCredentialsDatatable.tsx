import { useStore } from 'zustand';
import { Key } from 'lucide-react';

import { useUser } from '@/react/hooks/useUser';

import { Datatable } from '@@/datatables';
import { useSearchBarState } from '@@/datatables/SearchBar';
import { createPersistedStore } from '@@/datatables/types';

import { useGitCredentials } from '../gitCredential.service';

import { columns } from './columns';
import { GitCredentialsDatatableActions } from './GitCredentialsDatatableActions';

const storageKey = 'gitCredentials';

const settingsStore = createPersistedStore(storageKey);

export function GitCredentialsDatatable() {
  const { user } = useUser();
  const gitCredentialsQuery = useGitCredentials(user.Id);

  const settings = useStore(settingsStore);

  const [search, setSearch] = useSearchBarState(storageKey);
  return (
    <Datatable
      dataset={gitCredentialsQuery.data || []}
      columns={columns}
      initialPageSize={settings.pageSize}
      onPageSizeChange={settings.setPageSize}
      initialSortBy={settings.sortBy}
      onSortByChange={settings.setSortBy}
      searchValue={search}
      onSearchChange={setSearch}
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
