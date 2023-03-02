import { Select } from '@@/form-components/ReactSelect';

export type AddOnOption = {
  name: string;
};

interface Props {
  value: AddOnOption[];
  onChange(value: readonly AddOnOption[]): void;
  options: AddOnOption[];
}

export function Microk8sAddOnSelector({ value, onChange, options }: Props) {
  return (
    <Select
      isMulti
      getOptionLabel={(option) => option.name}
      getOptionValue={(option) => option.name}
      options={options}
      value={value}
      closeMenuOnSelect={false}
      onChange={onChange}
      placeholder="Select one or more addons"
    />
  );
}
