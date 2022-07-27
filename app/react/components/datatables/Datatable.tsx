import {
  useTable,
  useFilters,
  useGlobalFilter,
  useSortBy,
  usePagination,
  Column,
  TableInstance,
} from 'react-table';
import { ReactNode } from 'react';
import { useRowSelectColumn } from '@lineup-lite/hooks';

import { PaginationControls } from '@@/PaginationControls';
import { Table } from '@@/datatables/Table';
import { multiple } from '@@/datatables/filter-types';
import { SearchBar, useSearchBarState } from '@@/datatables/SearchBar';
import { SelectedRowsCount } from '@@/datatables/SelectedRowsCount';
import { TableSettingsProvider } from '@@/datatables/useTableSettings';
import { useRowSelect } from '@@/datatables/useRowSelect';
import { ColumnVisibilityMenu } from '@@/datatables/ColumnVisibilityMenu';
import {
  PaginationTableSettings,
  SettableColumnsTableSettings,
  SortableTableSettings,
} from '@@/datatables/types';

interface DefaultTableSettings
  extends SortableTableSettings,
    PaginationTableSettings,
    SettableColumnsTableSettings {
  setHiddenColumns: (hiddenColumns: string[]) => void;
  setSortBy: (id: string, desc: boolean) => void;
  setPageSize: (size: number) => void;
}

interface Props<
  D extends Record<string, unknown>,
  TSettings extends DefaultTableSettings
> {
  dataset: D[];
  storageKey: string;
  columns: Column<D>[];
  renderTableSettings?(instance: TableInstance<D>): ReactNode;
  renderTableActions?(selectedRows: D[]): ReactNode;
  settingsStore: TSettings;
  disableSelect?: boolean;
  hidableColumns?: string[];
  getRowId?(row: D): string;
  isRowSelectable?(row: D): boolean;
  emptyContentLabel?: string;
  titleOptions: {
    title: string;
    icon: string;
  };
}

export function Datatable<
  D extends Record<string, unknown>,
  TSettings extends DefaultTableSettings
>({
  columns,
  dataset,
  storageKey,
  renderTableSettings,
  renderTableActions,
  settingsStore,
  disableSelect,
  hidableColumns = [],
  getRowId = defaultGetRowId,
  isRowSelectable = () => true,
  titleOptions,
  emptyContentLabel,
}: Props<D, TSettings>) {
  const [searchBarValue, setSearchBarValue] = useSearchBarState(storageKey);

  const tableInstance = useTable<D>(
    {
      defaultCanFilter: false,
      columns,
      data: dataset,
      filterTypes: { multiple },
      initialState: {
        pageSize: settingsStore.pageSize || 10,
        hiddenColumns: settingsStore.hiddenColumns,
        sortBy: [settingsStore.sortBy],
        globalFilter: searchBarValue,
      },
      isRowSelectable,
      autoResetSelectedRows: false,
      getRowId,
      stateReducer: (newState, action) => {
        switch (action.type) {
          case 'setGlobalFilter':
            setSearchBarValue(action.filterValue);
            break;
          case 'toggleSortBy':
            settingsStore.setSortBy(action.columnId, action.desc);
            break;
          case 'setPageSize':
            settingsStore.setPageSize(action.pageSize);
            break;
          default:
            break;
        }
        return newState;
      },
    },
    useFilters,
    useGlobalFilter,
    useSortBy,
    usePagination,
    useRowSelect,
    !disableSelect ? useRowSelectColumn : emptyPlugin
  );

  const {
    selectedFlatRows,
    getTableProps,
    getTableBodyProps,
    headerGroups,
    page,
    prepareRow,
    gotoPage,
    setPageSize,
    setGlobalFilter,
    state: { pageIndex, pageSize },
  } = tableInstance;

  const tableProps = getTableProps();
  const tbodyProps = getTableBodyProps();

  const selectedItems = selectedFlatRows.map((row) => row.original);

  const columnsToHide = tableInstance.allColumns.filter((colInstance) =>
    hidableColumns?.includes(colInstance.id)
  );

  return (
    <div className="row">
      <div className="col-sm-12">
        <TableSettingsProvider defaults={settingsStore} storageKey={storageKey}>
          <Table.Container>
            <Table.Title label={titleOptions.title} icon={titleOptions.icon}>
              <Table.TitleActions>
                {hidableColumns.length > 0 && (
                  <ColumnVisibilityMenu<D>
                    columns={columnsToHide}
                    onChange={(hiddenColumns) => {
                      settingsStore.setHiddenColumns(hiddenColumns);
                      tableInstance.setHiddenColumns(hiddenColumns);
                    }}
                    value={settingsStore.hiddenColumns}
                  />
                )}

                {!!renderTableSettings && renderTableSettings(tableInstance)}
              </Table.TitleActions>
            </Table.Title>
            {renderTableActions && (
              <Table.Actions>{renderTableActions(selectedItems)}</Table.Actions>
            )}
            <SearchBar value={searchBarValue} onChange={setGlobalFilter} />
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
                    <Table.HeaderRow<D>
                      key={key}
                      className={className}
                      role={role}
                      style={style}
                      headers={headerGroup.headers}
                    />
                  );
                })}
              </thead>
              <tbody
                className={tbodyProps.className}
                role={tbodyProps.role}
                style={tbodyProps.style}
              >
                <Table.Content<D>
                  rows={page}
                  isLoading={false}
                  prepareRow={prepareRow}
                  emptyContent={emptyContentLabel}
                  renderRow={(row, { key, className, role, style }) => (
                    <Table.Row<D>
                      cells={row.cells}
                      key={key}
                      className={className}
                      role={role}
                      style={style}
                    />
                  )}
                />
              </tbody>
            </Table>
            <Table.Footer>
              <SelectedRowsCount value={selectedFlatRows.length} />
              <PaginationControls
                showAll
                pageLimit={pageSize}
                page={pageIndex + 1}
                onPageChange={(p) => gotoPage(p - 1)}
                totalCount={dataset.length}
                onPageLimitChange={setPageSize}
              />
            </Table.Footer>
          </Table.Container>
        </TableSettingsProvider>
      </div>
    </div>
  );
}

function defaultGetRowId<D extends Record<string, unknown>>(row: D): string {
  if (row.id && (typeof row.id === 'string' || typeof row.id === 'number')) {
    return row.id.toString();
  }

  if (row.Id && (typeof row.Id === 'string' || typeof row.Id === 'number')) {
    return row.Id.toString();
  }

  if (row.ID && (typeof row.ID === 'string' || typeof row.ID === 'number')) {
    return row.ID.toString();
  }

  return '';
}

function emptyPlugin() {}

emptyPlugin.pluginName = 'emptyPlugin';
