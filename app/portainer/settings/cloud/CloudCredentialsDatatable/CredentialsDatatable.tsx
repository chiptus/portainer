import { useEffect } from 'react';
import { useRowSelectColumn } from '@lineup-lite/hooks';
import {
  useTable,
  useSortBy,
  useFilters,
  useGlobalFilter,
  usePagination,
} from 'react-table';

import {
  TableActions,
  TableContainer,
  TableHeaderRow,
  TableRow,
  TableTitle,
} from '@/portainer/components/datatables/components';
import {
  SearchBar,
  useSearchBarState,
} from '@/portainer/components/datatables/components/SearchBar';
import { multiple } from '@/portainer/components/datatables/components/filter-types';
import { useTableSettings } from '@/portainer/components/datatables/components/useTableSettings';
import { useDebounce } from '@/portainer/hooks/useDebounce';
import { Table } from '@/portainer/components/datatables/components/Table';
import { TableFooter } from '@/portainer/components/datatables/components/TableFooter';
import { SelectedRowsCount } from '@/portainer/components/datatables/components/SelectedRowsCount';
import { PaginationControls } from '@/portainer/components/pagination-controls';
import { Checkbox } from '@/portainer/components/form-components/Checkbox';
import { useRowSelect } from '@/portainer/components/datatables/components/useRowSelect';

import { Credential, CredentialTableSettings } from '../types';

import { CredentialsDatatableActions } from './CredentialsDatatableActions';
import { useColumns } from './columns';

interface Props {
  storageKey: string;
  dataset: Credential[];
  isLoading: boolean;
}

export function CredentialsDatatable({
  storageKey,
  dataset,
  isLoading,
}: Props) {
  const { settings, setTableSettings } =
    useTableSettings<CredentialTableSettings>();
  const columns = useColumns();
  const [searchBarValue, setSearchBarValue] = useSearchBarState(storageKey);
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
  } = useTable<Credential>(
    {
      defaultCanFilter: false,
      columns,
      data: dataset,
      filterTypes: { multiple },
      initialState: {
        pageSize: settings.pageSize || 10,
        sortBy: [settings.sortBy],
        globalFilter: searchBarValue,
      },
      isRowSelectable() {
        return true;
      },
      autoResetSelectedRows: false,
      getRowId(row: Credential) {
        return String(row.id);
      },
      selectCheckboxComponent: Checkbox,
    },
    useFilters,
    useGlobalFilter,
    useSortBy,
    usePagination,
    useRowSelect,
    useRowSelectColumn
  );

  const debouncedSearchValue = useDebounce(searchBarValue);

  const tableProps = getTableProps();
  const tbodyProps = getTableBodyProps();

  useEffect(() => {
    setGlobalFilter(debouncedSearchValue);
  }, [debouncedSearchValue, setGlobalFilter]);

  return (
    <TableContainer>
      <TableTitle icon="fa-cloud" label="Cloud Provider Credentials" />

      <TableActions>
        <CredentialsDatatableActions
          selectedItems={selectedFlatRows.map((row) => row.original)}
        />
      </TableActions>
      <SearchBar
        value={searchBarValue}
        onChange={(value: string) => setSearchBarValue(value)}
        dataCy="credentials-searchBar"
      />
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
              <TableHeaderRow<Credential>
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
          {isLoading && (
            <tr>
              <td
                colSpan={columns.length + 1}
                className="text-center text-muted"
                data-cy="credentials-loading"
              >
                Loading...
              </td>
            </tr>
          )}
          {page.length === 0 && !isLoading && (
            <tr>
              <td
                colSpan={columns.length + 1}
                className="text-center text-muted"
                data-cy="credentials-noneAvailable"
              >
                No credential available.
              </td>
            </tr>
          )}
          {page.length >= 1 &&
            page.map((row) => {
              prepareRow(row);
              const { key, className, role, style } = row.getRowProps();
              return (
                <TableRow<Credential>
                  cells={row.cells}
                  key={key}
                  className={className}
                  role={role}
                  style={style}
                />
              );
            })}
        </tbody>
      </Table>

      <TableFooter>
        <SelectedRowsCount value={selectedFlatRows.length} />
        <PaginationControls
          showAll
          pageLimit={pageSize}
          page={pageIndex + 1}
          onPageChange={(p) => gotoPage(p - 1)}
          totalCount={dataset.length}
          onPageLimitChange={handlePageSizeChange}
        />
      </TableFooter>
    </TableContainer>
  );

  function handleSortChange(id: string, desc: boolean) {
    setTableSettings((settings) => ({
      ...settings,
      sortBy: { id, desc },
    }));
  }

  function handlePageSizeChange(pageSize: number) {
    setPageSize(pageSize);
    setTableSettings((settings) => ({ ...settings, pageSize }));
  }
}
