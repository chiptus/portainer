import { Fragment, useEffect } from 'react';
import {
  useExpanded,
  useFilters,
  useGlobalFilter,
  usePagination,
  useSortBy,
  useTable,
} from 'react-table';
import { useRowSelectColumn } from '@lineup-lite/hooks';

import { PaginationControls } from '@/portainer/components/pagination-controls';
import {
  Table,
  TableActions,
  TableContainer,
  TableHeaderRow,
  TableRow,
  TableTitle,
  TableSettingsMenu,
  TableTitleActions,
} from '@/portainer/components/datatables/components';
import { multiple } from '@/portainer/components/datatables/components/filter-types';
import { useTableSettings } from '@/portainer/components/datatables/components/useTableSettings';
import { useDebounce } from '@/portainer/hooks/useDebounce';
import {
  SearchBar,
  useSearchBarState,
} from '@/portainer/components/datatables/components/SearchBar';
import { useRowSelect } from '@/portainer/components/datatables/components/useRowSelect';
import { TableFooter } from '@/portainer/components/datatables/components/TableFooter';
import { SelectedRowsCount } from '@/portainer/components/datatables/components/SelectedRowsCount';
import { Job } from '@/nomad/types';
import { TableContent } from '@/portainer/components/datatables/components/TableContent';
import { useRepeater } from '@/portainer/components/datatables/components/useRepeater';

import { JobsTableSettings } from './types';
import { TasksDatatable } from './TasksDatatable';
import { useColumns } from './columns';
import { JobsDatatableSettings } from './JobsDatatableSettings';

export interface JobsDatatableProps {
  jobs: Job[];
  refreshData: () => Promise<void>;
  isLoading?: boolean;
}

export function JobsDatatable({
  jobs,
  refreshData,
  isLoading,
}: JobsDatatableProps) {
  const { settings, setTableSettings } = useTableSettings<JobsTableSettings>();
  const [searchBarValue, setSearchBarValue] = useSearchBarState('jobs');
  const columns = useColumns();
  const debouncedSearchValue = useDebounce(searchBarValue);
  useRepeater(settings.autoRefreshRate, refreshData);

  const {
    getTableProps,
    getTableBodyProps,
    headerGroups,
    page,
    prepareRow,
    selectedFlatRows,
    gotoPage,
    setPageSize,
    setGlobalFilter,
    state: { pageIndex, pageSize },
  } = useTable<Job>(
    {
      defaultCanFilter: false,
      columns,
      data: jobs,
      filterTypes: { multiple },
      initialState: {
        pageSize: settings.pageSize || 10,
        sortBy: [settings.sortBy],
        globalFilter: searchBarValue,
      },
      isRowSelectable() {
        return false;
      },
      autoResetExpanded: false,
      autoResetSelectedRows: false,
      selectColumnWidth: 5,
      getRowId(job, relativeIndex) {
        return `${job.ID}-${relativeIndex}`;
      },
    },
    useFilters,
    useGlobalFilter,
    useSortBy,
    useExpanded,
    usePagination,
    useRowSelect,
    useRowSelectColumn
  );

  useEffect(() => {
    setGlobalFilter(debouncedSearchValue);
  }, [debouncedSearchValue, setGlobalFilter]);

  const tableProps = getTableProps();
  const tbodyProps = getTableBodyProps();

  return (
    <TableContainer>
      <TableTitle icon="fa-cubes" label="Nomad Jobs">
        <TableTitleActions>
          <TableSettingsMenu>
            <JobsDatatableSettings />
          </TableSettingsMenu>
        </TableTitleActions>
      </TableTitle>

      <TableActions />

      <SearchBar value={searchBarValue} onChange={handleSearchBarChange} />

      <Table
        className={tableProps.className}
        role={tableProps.role}
        style={tableProps.style}
      >
        <thead>
          {headerGroups.map((headerGroup) => {
            const { key, className, role, style } =
              headerGroup.getHeaderGroupProps();

            return (
              <TableHeaderRow<Job>
                key={key}
                className={className}
                role={role}
                style={style}
                headers={headerGroup.headers}
                onSortChange={handleSortChange}
              />
            );
          })}
        </thead>
        <tbody
          className={tbodyProps.className}
          role={tbodyProps.role}
          style={tbodyProps.style}
        >
          <TableContent
            rows={page}
            prepareRow={prepareRow}
            isLoading={isLoading}
            emptyContent="No jobs found"
            renderRow={(row, { key, className, role, style }) => (
              <Fragment key={key}>
                <TableRow<Job>
                  cells={row.cells}
                  key={key}
                  className={className}
                  role={role}
                  style={style}
                />

                {row.isExpanded && (
                  <tr>
                    <td />
                    <td colSpan={row.cells.length - 1}>
                      <TasksDatatable data={row.original.Tasks} />
                    </td>
                  </tr>
                )}
              </Fragment>
            )}
          />
        </tbody>
      </Table>

      <TableFooter>
        <SelectedRowsCount value={selectedFlatRows.length} />
        <PaginationControls
          showAll
          pageLimit={pageSize}
          page={pageIndex + 1}
          onPageChange={(p) => gotoPage(p - 1)}
          totalCount={jobs.length}
          onPageLimitChange={handlePageSizeChange}
        />
      </TableFooter>
    </TableContainer>
  );

  function handlePageSizeChange(pageSize: number) {
    setPageSize(pageSize);
    setTableSettings((settings) => ({ ...settings, pageSize }));
  }

  function handleSearchBarChange(value: string) {
    setSearchBarValue(value);
  }

  function handleSortChange(id: string, desc: boolean) {
    setTableSettings((settings) => ({
      ...settings,
      sortBy: { id, desc },
    }));
  }
}
