import { Select } from '@@/form-components/ReactSelect';

import { Microk8sAddOn } from '../types';

interface Props {
  value: Microk8sAddOn[];
  onChange(value: readonly Microk8sAddOn[]): void;
  options: Microk8sAddOn[];
}

export function Microk8sAddOnSelector({ value, onChange, options }: Props) {
  return (
    <Select
      isMulti
      getOptionLabel={(option) => option.Name}
      getOptionValue={(option) => option.Name}
      options={options}
      value={value}
      closeMenuOnSelect={false}
      onChange={onChange}
      placeholder="Select one or more addons"
    />
  );
}
