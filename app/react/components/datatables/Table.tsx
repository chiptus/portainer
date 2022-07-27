import clsx from 'clsx';
import { PropsWithChildren } from 'react';
import { TableProps } from 'react-table';

import { TableActions } from './TableActions';
import { TableContainer, useTableContext } from './TableContainer';
import { TableContent } from './TableContent';
import { TableFooter } from './TableFooter';
import { TableHeaderCell } from './TableHeaderCell';
import { TableHeaderRow } from './TableHeaderRow';
import { TableRow } from './TableRow';
import { TableSettingsMenu } from './TableSettingsMenu';
import { TableTitle } from './TableTitle';
import { TableTitleActions } from './TableTitleActions';

function Table({
  children,
  className,
  role,
  style,
}: PropsWithChildren<TableProps>) {
  useTableContext();

  return (
    <div className="table-responsive">
      <table
        className={clsx(
          'table table-hover table-filters nowrap-cells',
          className
        )}
        role={role}
        style={style}
      >
        {children}
      </table>
    </div>
  );
}

interface SubComponents {
  Container: typeof TableContainer;
  Actions: typeof TableActions;
  TitleActions: typeof TableTitleActions;
  HeaderCell: typeof TableHeaderCell;
  SettingsMenu: typeof TableSettingsMenu;
  Title: typeof TableTitle;
  Row: typeof TableRow;
  HeaderRow: typeof TableHeaderRow;
  Content: typeof TableContent;
  Footer: typeof TableFooter;
}

const ExportedTable: typeof Table & SubComponents = Table as typeof Table &
  SubComponents;

Table.Actions = TableActions;
Table.TitleActions = TableTitleActions;
Table.Container = TableContainer;
Table.HeaderCell = TableHeaderCell;
Table.SettingsMenu = TableSettingsMenu;
Table.Title = TableTitle;
Table.Row = TableRow;
Table.HeaderRow = TableHeaderRow;
Table.Content = TableContent;
Table.Footer = TableFooter;

export { ExportedTable as Table };
