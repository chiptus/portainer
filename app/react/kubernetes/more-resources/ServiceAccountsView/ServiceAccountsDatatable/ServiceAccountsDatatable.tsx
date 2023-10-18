import { Trash2, User } from 'lucide-react';
import { useRouter } from '@uirouter/react';
import { useEffect, useMemo } from 'react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { Authorized, useAuthorizations } from '@/react/hooks/useUser';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { SystemResourceDescription } from '@/react/kubernetes/datatables/SystemResourceDescription';
import { useNamespacesQuery } from '@/react/kubernetes/namespaces/queries/useNamespacesQuery';
import { createStore } from '@/react/kubernetes/datatables/default-kube-datatable-store';

import { Datatable, TableSettingsMenu } from '@@/datatables';
import { confirmDelete } from '@@/modals/confirm';
import { Button } from '@@/buttons';
import { Link } from '@@/Link';
import { useTableState } from '@@/datatables/useTableState';

import { ServiceAccount } from '../types';
import { DefaultDatatableSettings } from '../../../datatables/DefaultDatatableSettings';

import { useColumns } from './columns';
import { useGetServiceAccountsQuery } from './queries/useGetServiceAccountsQuery';
import { useDeleteServiceAccountsMutation } from './queries/useDeleteServiceAccountsMutation';

const storageKey = 'serviceAccounts';
const settingsStore = createStore(storageKey);

export function ServiceAccountsDatatable() {
  const environmentId = useEnvironmentId();
  const tableState = useTableState(settingsStore, storageKey);
  const namespacesQuery = useNamespacesQuery(environmentId);
  const namespaceNames = Object.keys(namespacesQuery.data || {});
  const serviceAccountsQuery = useGetServiceAccountsQuery(
    environmentId,
    namespaceNames,
    {
      autoRefreshRate: tableState.autoRefreshRate * 1000,
      enabled: namespacesQuery.isSuccess,
    }
  );
  const router = useRouter();
  const columns = useColumns();
  const isAuthorisedToAddEdit = useAuthorizations(['K8sServiceAccountsW']);
  const filteredServiceAccounts = useMemo(
    () =>
      serviceAccountsQuery.data?.filter(
        (sa) => tableState.showSystemResources || !sa.isSystem
      ),
    [serviceAccountsQuery.data, tableState.showSystemResources]
  );

  useEffect(() => {
    if (!isAuthorisedToAddEdit) {
      router.stateService.go('kubernetes.dashboard');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthorisedToAddEdit]);

  return (
    <Datatable
      dataset={filteredServiceAccounts || []}
      columns={columns}
      settingsManager={tableState}
      isLoading={serviceAccountsQuery.isLoading}
      emptyContentLabel="No service accounts found"
      title="Service Accounts"
      titleIcon={User}
      getRowId={(row) => row.uid}
      isRowSelectable={(row) => !row.original.isSystem}
      disableSelect={!isAuthorisedToAddEdit}
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
    />
  );
}

interface SelectedServiceAccount {
  namespace: string;
  name: string;
}

type TableActionsProps = {
  selectedItems: ServiceAccount[];
};

function TableActions({ selectedItems }: TableActionsProps) {
  const environmentId = useEnvironmentId();
  const deleteServiceAccountsMutation =
    useDeleteServiceAccountsMutation(environmentId);
  const router = useRouter();

  async function handleRemoveClick(serviceAccounts: SelectedServiceAccount[]) {
    const confirmed = await confirmDelete(
      <>
        <p>Are you sure you want to delete the selected service account(s)?</p>
        <ul className="mt-2 max-h-96 list-inside overflow-hidden overflow-y-auto text-sm">
          {serviceAccounts.map((s, index) => (
            <li key={index}>
              {s.namespace}/{s.name}
            </li>
          ))}
        </ul>
      </>
    );
    if (!confirmed) {
      return null;
    }

    const payload: Record<string, string[]> = {};
    serviceAccounts.forEach((sa) => {
      payload[sa.namespace] = payload[sa.namespace] || [];
      payload[sa.namespace].push(sa.name);
    });

    deleteServiceAccountsMutation.mutate(
      { environmentId, data: payload },
      {
        onSuccess: () => {
          notifySuccess(
            'Service account(s) successfully removed',
            serviceAccounts.map((sa) => `${sa.namespace}/${sa.name}`).join(', ')
          );
          router.stateService.reload();
        },
        onError: (error) => {
          notifyError(
            'Unable to delete service account(s)',
            error as Error,
            serviceAccounts.map((sa) => `${sa.namespace}/${sa.name}`).join(', ')
          );
        },
      }
    );
    return serviceAccounts;
  }

  return (
    <Authorized authorizations="K8sServiceAccountsW">
      <Button
        className="btn-wrapper"
        color="dangerlight"
        disabled={selectedItems.length === 0}
        onClick={() => handleRemoveClick(selectedItems)}
        icon={Trash2}
      >
        Remove
      </Button>

      <Link
        to="kubernetes.deploy"
        className="ml-1"
        params={{ referrer: 'kubernetes.moreResources.serviceAccounts' }}
      >
        <Button className="btn-wrapper" color="primary" icon="plus">
          Create from manifest
        </Button>
      </Link>
    </Authorized>
  );
}
