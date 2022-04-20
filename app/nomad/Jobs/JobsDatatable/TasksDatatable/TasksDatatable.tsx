import { useFilters, usePagination, useSortBy, useTable } from 'react-table';
import { useState } from 'react';

import {
  Table,
  TableContainer,
  TableHeaderRow,
  TableRow,
} from '@/portainer/components/datatables/components';
import { InnerDatatable } from '@/portainer/components/datatables/components/InnerDatatable';
import { Task } from '@/nomad/types';

import { useColumns } from './columns';

export interface TasksTableProps {
  data: Task[];
}

export function TasksDatatable({ data }: TasksTableProps) {
  const columns = useColumns();
  const [sortBy, setSortBy] = useState({ id: 'taskName', desc: false });

  const { getTableProps, getTableBodyProps, headerGroups, page, prepareRow } =
    useTable<Task>(
      {
        columns,
        data,
        initialState: {
          sortBy: [sortBy],
        },
      },
      useFilters,
      useSortBy,
      usePagination
    );

  const tableProps = getTableProps();
  const tbodyProps = getTableBodyProps();

  return (
    <InnerDatatable>
      <TableContainer>
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
                <TableHeaderRow<Task>
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
            {data.length > 0 ? (
              page.map((row) => {
                prepareRow(row);
                const { key, className, role, style } = row.getRowProps();

                return (
                  <TableRow<Task>
                    key={key}
                    cells={row.cells}
                    className={className}
                    role={role}
                    style={style}
                  />
                );
              })
            ) : (
              <tr>
                <td colSpan={5} className="text-center text-muted">
                  no tasks
                </td>
              </tr>
            )}
          </tbody>
        </Table>
      </TableContainer>
    </InnerDatatable>
  );

  function handleSortChange(id: string, desc: boolean) {
    setSortBy({ id, desc });
  }
}
