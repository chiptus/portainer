import { FormikHandlers } from 'formik';

import { RegistryId } from '@/react/portainer/registries/types/registry';

import { FormControl } from '@@/form-components/FormControl';
import { Select } from '@@/form-components/ReactSelect';
import { Option } from '@@/form-components/Input/Select';

interface Props {
  errorMessage: string;
  onBlur: FormikHandlers['handleBlur'];
  value: RegistryId;
  onChange(value: RegistryId): void;
  disabled?: boolean;
  registries: Option<RegistryId>[];
}

export function RegistrySelector({
  errorMessage,
  onBlur,
  value,
  onChange,
  disabled = false,
  registries,
}: Props) {
  return (
    <FormControl
      label="Registry"
      errors={errorMessage}
      inputId="registry-select"
    >
      <Select
        name="registryId"
        onBlur={onBlur}
        onChange={(option) => {
          onChange(option?.value || 0);
        }}
        value={registries.find((registry) => registry.value === value)}
        inputId="registry-select"
        options={registries}
        getOptionLabel={(registry) => registry.label}
        getOptionValue={(registry) => registry.value.toString()}
        isDisabled={disabled}
      />
    </FormControl>
  );
}
