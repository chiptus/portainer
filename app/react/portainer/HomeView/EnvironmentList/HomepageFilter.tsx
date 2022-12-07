import { components, OptionProps } from 'react-select';

import { useLocalStorage } from '@/react/hooks/useLocalStorage';

import { MultiSelect } from '@@/form-components/PortainerSelect';

import { Filter } from './types';

interface Props<TValue = number> {
  filterOptions?: Array<Filter<TValue>>;
  onChange: (filterOptions: TValue[]) => void;
  placeholder: string;
  value: TValue[];
}

function Option<TValue = number>(props: OptionProps<TValue>) {
  const { isSelected, label } = props;
  return (
    <div>
      <components.Option
        // eslint-disable-next-line react/jsx-props-no-spreading
        {...props}
      >
        <input type="checkbox" checked={isSelected} onChange={() => null} />{' '}
        <label>{label}</label>
      </components.Option>
    </div>
  );
}

export function HomepageFilter<TValue = number>({
  filterOptions = [],
  onChange,
  placeholder,
  value,
}: Props<TValue>) {
  return (
    <MultiSelect
      placeholder={placeholder}
      options={filterOptions}
      value={value}
      components={{ Option }}
      onChange={(option) => onChange([...option])}
    />
  );
}

export function useHomePageFilter<T>(
  key: string,
  defaultValue: T
): [T, (value: T) => void] {
  const filterKey = keyBuilder(key);
  return useLocalStorage(filterKey, defaultValue, sessionStorage);
}

function keyBuilder(key: string) {
  return `datatable_home_filter_type_${key}`;
}
