// import { useState } from 'react';
import { Plus, Puzzle, Trash2 } from 'lucide-react';

import { notifySuccess } from '@/portainer/services/notifications';
import { withLimitToBE } from '@/react/hooks/useLimitToBE';
// import { withSortQuery } from '@/react/common/api/sort.types';

import { confirmDelete } from '@@/modals/confirm';
import { Datatable } from '@@/datatables';
import { PageHeader } from '@@/PageHeader';
import { Button } from '@@/buttons';
import { Link } from '@@/Link';
import { useTableState } from '@@/datatables/useTableState';

import { useList } from '../queries/list/list';
import { EdgeConfiguration } from '../types';
import { useRemoveMutation } from '../queries/useRemoveMutation';

import { columns } from './columns';
import { createStore } from './datatable-store';

const storageKey = 'edge-configurations-list';
const settingsStore = createStore(storageKey);

export default withLimitToBE(ListView);

// const initialPage = 0;

export function ListView() {
  // const [page, setPage] = useState(initialPage);
  const tableState = useTableState(settingsStore, storageKey);
  const { data: configurations, isLoading } = useList({
    // page,
    // pageLimit: tableState.pageSize,
    // search: tableState.search,
    // ...withSortQuery(tableState.sortBy, sortOptions),
  });

  // const pageCount = Math.ceil(totalCount / tableState.pageSize);

  if (!configurations) {
    return null;
  }

  return (
    <>
      <PageHeader
        title="Edge configurations"
        breadcrumbs="Edge configurations"
        reload
      />

      <Datatable
        dataset={configurations}
        columns={columns}
        settingsManager={tableState}
        title="Edge configurations list"
        titleIcon={Puzzle}
        emptyContentLabel="No edge configurations found"
        renderTableActions={(selectedRows) => (
          <TableActions selectedRows={selectedRows} />
        )}
        isLoading={isLoading}
        // totalCount={totalCount}
        // pageCount={pageCount}
        // onPageChange={setPage}
        // onSearchChange={() => setPage(initialPage)}
      />
    </>
  );
}

function TableActions({ selectedRows }: { selectedRows: EdgeConfiguration[] }) {
  const removeMutation = useRemoveMutation();
  return (
    <>
      <Button
        icon={Trash2}
        color="dangerlight"
        onClick={() => handleRemove()}
        disabled={selectedRows.length === 0}
      >
        Remove
      </Button>

      <Link to=".create">
        <Button icon={Plus}>Add configuration</Button>
      </Link>
    </>
  );

  async function handleRemove() {
    const confirmed = await confirmDelete(
      'Do you want to remove the selected configuration(s) from all devices within the edge groups corresponding to these configuration(s)?'
    );
    if (!confirmed) {
      return;
    }

    removeMutation.mutate(selectedRows, {
      onSuccess: () => {
        notifySuccess('Success', 'Configurations successfully removed');
      },
    });
  }
}
