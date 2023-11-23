import { Datatable as GenericDatatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { columns } from './columns';
import { TableActions } from './TableActions';
import { useEdgeAdmins } from './useEdgeAdmins';

const storageKey = 'edge-admins-list';

const settingsStore = createPersistedStore(storageKey, 'name');

export function Datatable() {
  const tableState = useTableState(settingsStore, storageKey);
  const { data, isLoading } = useEdgeAdmins();

  return (
    <GenericDatatable
      noWidget
      settingsManager={tableState}
      columns={columns}
      dataset={data}
      title="Edge administrators list"
      emptyContentLabel="No edge administrator found"
      renderTableActions={(selectedRows) => (
        <TableActions selectedRows={selectedRows} />
      )}
      isLoading={isLoading}
    />
  );
}
