import { useUser } from '@/portainer/hooks/useUser';

import { TableSettingsProvider } from '@@/datatables/useTableSettings';

import { useGitCredentials } from '../gitCredential.service';

import { GitCredentialsDatatable } from './GitCredentialsDatatable';

export default function CredentialsDatatableContainer() {
  const defaultSettings = {
    autoRefreshRate: 0,
    pageSize: 10,
    sortBy: { id: 'state', desc: false },
  };
  const currentUser = useUser();

  const storageKey = 'gitCredentials';

  const gitCredentialsQuery = useGitCredentials(currentUser.user.Id);

  const gitCredentials = gitCredentialsQuery.data || [];

  return (
    <TableSettingsProvider defaults={defaultSettings} storageKey={storageKey}>
      <GitCredentialsDatatable
        storageKey={storageKey}
        dataset={gitCredentials}
        isLoading={gitCredentialsQuery.isLoading}
      />
    </TableSettingsProvider>
  );
}
