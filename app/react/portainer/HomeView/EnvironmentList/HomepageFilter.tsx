import { components, OptionProps } from 'react-select';

import { Select as ReactSelect } from '@@/form-components/ReactSelect';

import { Filter } from './types';

interface Props<TValue = number> {
  filterOptions?: Array<Filter<TValue>>;
  onChange: (filterOptions: TValue[]) => void;
  placeholder: string;
  value: TValue[];
}

function Option<TValue = number>(props: OptionProps<Filter<TValue>, true>) {
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
  const selectedValue = filterOptions.filter((option) =>
    value.includes(option.value)
  );

  return (
    <ReactSelect
      closeMenuOnSelect={false}
      placeholder={placeholder}
      options={filterOptions}
      value={selectedValue}
      isMulti
      components={{ Option }}
      onChange={(option) => onChange(option.map((o) => o.value))}
    />
  );
}
