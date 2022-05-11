import { TableSettingsProvider } from '@/portainer/components/datatables/components/useTableSettings';

import { useCloudCredentials } from '../cloudSettings.service';

import { CredentialsDatatable } from './CredentialsDatatable';

export function CredentialsDatatableContainer() {
  const defaultSettings = {
    autoRefreshRate: 0,
    pageSize: 10,
    sortBy: { id: 'state', desc: false },
  };
  const storageKey = 'cloudCredentials';

  const cloudCredentialsQuery = useCloudCredentials();

  const credentials = cloudCredentialsQuery.data || [];

  return (
    <TableSettingsProvider defaults={defaultSettings} storageKey={storageKey}>
      <CredentialsDatatable
        storageKey={storageKey}
        dataset={credentials}
        isLoading={cloudCredentialsQuery.isLoading}
      />
    </TableSettingsProvider>
  );
}
