import { Select } from '@@/form-components/ReactSelect';

export type AddOnOption = {
  name: string;
};

interface Props {
  value: AddOnOption[];
  onChange(value: readonly AddOnOption[]): void;
  options: AddOnOption[];
  disabled?: boolean;
}

export function Microk8sAddOnSelector({
  value,
  onChange,
  options,
  disabled,
}: Props) {
  return (
    <Select
      isMulti
      getOptionLabel={(option) => option.name}
      getOptionValue={(option) => option.name}
      options={options}
      value={value}
      isDisabled={disabled}
      closeMenuOnSelect={false}
      onChange={onChange}
      placeholder="Select one or more addons"
    />
  );
}
