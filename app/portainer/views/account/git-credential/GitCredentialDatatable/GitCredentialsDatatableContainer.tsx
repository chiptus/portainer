import { useUser } from '@/portainer/hooks/useUser';
import { react2angular } from '@/react-tools/react2angular';

import { TableSettingsProvider } from '@@/datatables/useTableSettings';

import { useGitCredentials } from '../gitCredential.service';

import { GitCredentialsDatatable } from './GitCredentialsDatatable';

export function CredentialsDatatableContainer() {
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

export const gitCredentialsDatatable = react2angular(
  CredentialsDatatableContainer,
  []
);
