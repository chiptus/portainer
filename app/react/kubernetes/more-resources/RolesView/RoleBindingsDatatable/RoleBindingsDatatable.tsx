import { Plus, Trash2, Link as LinkIcon } from 'lucide-react';
import { useRouter } from '@uirouter/react';
import { Row } from '@tanstack/react-table';
import clsx from 'clsx';
import { useMemo } from 'react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useAuthorizations, Authorized } from '@/react/hooks/useUser';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { SystemResourceDescription } from '@/react/kubernetes/datatables/SystemResourceDescription';
import { createStore } from '@/react/kubernetes/datatables/default-kube-datatable-store';
import { useNamespacesQuery } from '@/react/kubernetes/namespaces/queries/useNamespacesQuery';

import { confirmDelete } from '@@/modals/confirm';
import { Datatable, Table, TableSettingsMenu } from '@@/datatables';
import { Button, LoadingButton } from '@@/buttons';
import { Link } from '@@/Link';
import { useTableState } from '@@/datatables/useTableState';

import { DefaultDatatableSettings } from '../../../datatables/DefaultDatatableSettings';

import { RoleBinding } from './types';
import { columns } from './columns';
import { useRoleBindingsQuery } from './queries/useRoleBindingsQuery';
import { useDeleteRoleBindingsMutation } from './queries/useDeleteRoleBindingsMutation';

const storageKey = 'roleBindings';
const settingsStore = createStore(storageKey);

export function RoleBindingsDatatable() {
  const environmentId = useEnvironmentId();
  const tableState = useTableState(settingsStore, storageKey);
  const namespacesQuery = useNamespacesQuery(environmentId);
  const namespaceNames = Object.keys(namespacesQuery.data || {});
  const roleBindingsQuery = useRoleBindingsQuery(
    environmentId,
    namespaceNames,
    {
      autoRefreshRate: tableState.autoRefreshRate * 1000,
      enabled: namespacesQuery.isSuccess,
    }
  );
  const filteredRoleBindings = useMemo(
    () =>
      roleBindingsQuery.data?.filter(
        (rb) => tableState.showSystemResources || !rb.isSystem
      ),
    [roleBindingsQuery.data, tableState.showSystemResources]
  );

  const isAuthorisedToAddEdit = useAuthorizations(['K8sRoleBindingsW']);

  return (
    <Datatable
      dataset={filteredRoleBindings || []}
      columns={columns}
      settingsManager={tableState}
      isLoading={roleBindingsQuery.isLoading}
      emptyContentLabel="No role bindings found"
      title="Role Bindings"
      titleIcon={LinkIcon}
      getRowId={(row) => row.uid}
      isRowSelectable={(row) => !row.original.isSystem}
      renderTableActions={(selectedRows) => (
        <TableActions selectedItems={selectedRows} />
      )}
      renderTableSettings={() => (
        <TableSettingsMenu>
          <DefaultDatatableSettings settings={tableState} />
        </TableSettingsMenu>
      )}
      description={
        <SystemResourceDescription
          showSystemResources={tableState.showSystemResources}
        />
      }
      disableSelect={!isAuthorisedToAddEdit}
      renderRow={renderRow}
    />
  );
}

// needed to apply custom styling to the row and not globally required in the AC's for this ticket.
function renderRow(row: Row<RoleBinding>, highlightedItemId?: string) {
  return (
    <Table.Row<RoleBinding>
      cells={row.getVisibleCells()}
      className={clsx('[&>td]:!py-4 [&>td]:!align-top', {
        active: highlightedItemId === row.id,
      })}
    />
  );
}

interface SelectedRole {
  namespace: string;
  name: string;
}

type TableActionsProps = {
  selectedItems: RoleBinding[];
};

function TableActions({ selectedItems }: TableActionsProps) {
  const environmentId = useEnvironmentId();
  const deleteRoleBindingsMutation =
    useDeleteRoleBindingsMutation(environmentId);
  const router = useRouter();

  async function handleRemoveClick(roles: SelectedRole[]) {
    const confirmed = await confirmDelete(
      <>
        <p>Are you sure you want to delete the selected roles(s)?</p>
        <ul className="mt-2 max-h-96 list-inside overflow-hidden overflow-y-auto text-sm">
          {roles.map((r, index) => (
            <li key={index}>
              {r.namespace}/{r.name}
            </li>
          ))}
        </ul>
      </>
    );
    if (!confirmed) {
      return null;
    }

    const payload: Record<string, string[]> = {};
    roles.forEach((r) => {
      payload[r.namespace] = payload[r.namespace] || [];
      payload[r.namespace].push(r.name);
    });

    deleteRoleBindingsMutation.mutate(
      { environmentId, data: payload },
      {
        onSuccess: () => {
          notifySuccess(
            'Role binding(s) successfully removed',
            roles.map((r) => `${r.namespace}/${r.name}`).join(', ')
          );
          router.stateService.reload();
        },
        onError: (error) => {
          notifyError(
            'Unable to delete role bindings(s)',
            error as Error,
            roles.map((r) => `${r.namespace}/${r.name}`).join(', ')
          );
        },
      }
    );
    return roles;
  }

  return (
    <Authorized authorizations="K8sRoleBindingsW">
      <LoadingButton
        className="btn-wrapper"
        color="dangerlight"
        disabled={selectedItems.length === 0}
        onClick={() => handleRemoveClick(selectedItems)}
        icon={Trash2}
        isLoading={deleteRoleBindingsMutation.isLoading}
        loadingText="Removing role bindings..."
      >
        Remove
      </LoadingButton>

      <Link
        to="kubernetes.deploy"
        params={{
          referrer: 'kubernetes.moreResources.roleBindings',
          tab: 'roleBindings',
        }}
        className="ml-1"
      >
        <Button className="btn-wrapper" color="primary" icon={Plus}>
          Create from manifest
        </Button>
      </Link>
    </Authorized>
  );
}