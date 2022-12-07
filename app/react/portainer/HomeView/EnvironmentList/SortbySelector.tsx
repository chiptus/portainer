import clsx from 'clsx';

import { TableHeaderSortIcons } from '@@/datatables/TableHeaderSortIcons';
import { SingleSelect } from '@@/form-components/PortainerSelect';

import { Filter } from './types';
import styles from './SortbySelector.module.css';

interface Props {
  filterOptions: Filter<string>[];
  onChange: (filterOptions?: string) => void;
  value?: string;
  onDescendingChange: (desc: boolean) => void;
  placeHolder: string;
  sortByDescending: boolean;
}

export function SortSelector({
  filterOptions,
  onChange,
  onDescendingChange,
  placeHolder,
  sortByDescending,

  value,
}: Props) {
  const sorted = !!value;
  return (
    <div className={styles.sortByContainer}>
      <div className={styles.sortByElement}>
        <SingleSelect
          placeholder={placeHolder}
          options={filterOptions}
          onChange={(option) => onChange(option || undefined)}
          isClearable
          value={value}
        />
      </div>
      <div className={styles.sortByElement}>
        <button
          className={clsx(styles.sortButton, 'h-[34px]')}
          type="button"
          disabled={!sorted}
          onClick={(e) => {
            e.preventDefault();
            onDescendingChange(!sortByDescending);
          }}
        >
          <TableHeaderSortIcons sorted={sorted} descending={sortByDescending} />
        </button>
      </div>
    </div>
  );
}
