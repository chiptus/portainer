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
import { Tab, WidgetTabs, useCurrentTabIndex } from '@@/Widget/WidgetTabs';

import { useList } from '../queries/list/list';
import { EdgeConfiguration } from '../types';
import { useRemoveMutation } from '../queries/useRemoveMutation';

import { columns } from './columns';
import { createStore } from './datatable-store';

const storageKey = 'edge-configurations-list';
const settingsStore = createStore(storageKey);

export default withLimitToBE(ListView);

// const initialPage = 0;

const tabs: Tab[] = [
  {
    name: 'Configurations',
    widget: <div />,
    selectedTabParam: 'configurations',
  },
  {
    name: 'Secrets',
    widget: <div />,
    selectedTabParam: 'secrets',
  },
];

const categories = ['configuration', 'secret'];

export function ListView() {
  // const [page, setPage] = useState(initialPage);
  const tableState = useTableState(settingsStore, storageKey);

  const [currentTabIndex] = useCurrentTabIndex(tabs);

  const title = `${tabs[currentTabIndex].name} list`;

  const category = categories[currentTabIndex];

  const { data: configurations, isLoading } = useList({
    category,
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

      <WidgetTabs tabs={tabs} currentTabIndex={currentTabIndex} />

      <Datatable
        dataset={configurations}
        columns={columns}
        settingsManager={tableState}
        title={title}
        titleIcon={Puzzle}
        emptyContentLabel={`No edge ${category}s found`}
        renderTableActions={(selectedRows) => (
          <TableActions selectedRows={selectedRows} category={category} />
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

function TableActions({
  selectedRows,
  category,
}: {
  selectedRows: EdgeConfiguration[];
  category: string;
}) {
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

      <Link to=".create" params={{ category }}>
        <Button icon={Plus}>Add {category}</Button>
      </Link>
    </>
  );

  async function handleRemove() {
    const confirmed = await confirmDelete(
      `Do you want to remove the selected ${category}(s) from all devices within the edge groups corresponding to these ${category}(s)?`
    );
    if (!confirmed) {
      return;
    }

    removeMutation.mutate(selectedRows, {
      onSuccess: () => {
        const upCategory =
          category.slice(0, 1).toUpperCase() + category.slice(1);

        notifySuccess('Success', `${upCategory}s successfully removed`);
      },
    });
  }
}
