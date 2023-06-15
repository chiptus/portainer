import { usePaginationState } from '@/react/hooks/usePaginationLimitState';

import { Datatable as GenericDatatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { columns } from './columns';
import { Filter } from './Filter';
import { TableActions } from './TableActions';
import { useEnvironments } from './useEnvironments';

const storageKey = 'edge-devices-waiting-room';

const settingsStore = createPersistedStore(storageKey);

export function Datatable() {
  const { page, pageLimit, setPage } = usePaginationState(storageKey);
  const tableState = useTableState(settingsStore, storageKey);
  const {
    data: environments,
    totalCount,
    isLoading,
  } = useEnvironments({ page: page + 1, pageLimit: tableState.pageSize });

  const pageCount = Math.ceil(totalCount / pageLimit);

  return (
    <GenericDatatable
      settingsManager={tableState}
      columns={columns}
      dataset={environments}
      title="Edge Devices Waiting Room"
      emptyContentLabel="No Edge Devices found"
      renderTableActions={(selectedRows) => (
        <TableActions selectedRows={selectedRows} />
      )}
      isLoading={isLoading}
      totalCount={totalCount}
      pageCount={pageCount}
      onPageChange={setPage}
      description={<Filter />}
    />
  );
}
