import { ReactNode } from 'react';

import { IconProps } from '@@/Icon';

import { SearchBar } from './SearchBar';
import { Table } from './Table';

type Props = {
  title?: string;
  titleIcon?: IconProps['icon'];
  searchValue: string;
  description?: ReactNode;
  onSearchChange(value: string): void;
  renderTableSettings?(): ReactNode;
  renderTableActions?(): ReactNode;
};

export function DatatableHeader({
  onSearchChange,
  renderTableActions,
  renderTableSettings,
  searchValue,
  description,
  title,
  titleIcon,
}: Props) {
  if (!title) {
    return null;
  }

  return (
    <Table.Title label={title} icon={titleIcon} description={description}>
      <SearchBar value={searchValue} onChange={onSearchChange} />
      {renderTableActions && (
        <Table.Actions>{renderTableActions()}</Table.Actions>
      )}
      <Table.TitleActions>
        {!!renderTableSettings && renderTableSettings()}
      </Table.TitleActions>
    </Table.Title>
  );
}
