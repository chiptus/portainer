import { useField } from 'formik';
import { ChangeEvent, useMemo } from 'react';

import { CustomTemplate } from '@/react/portainer/settings/cloud/types';

import { FormControl } from '@@/form-components/FormControl';
import { Select, Option } from '@@/form-components/Input/Select';

import { useSetAvailableOption } from '../useSetAvailableOption';

interface Props {
  customTemplates: CustomTemplate[];
}

export function CustomTemplateSelector({ customTemplates }: Props) {
  const customTemplateOptions: Option<number>[] = useMemo(() => {
    if (customTemplates.length === 0) {
      return [{ value: 0, label: 'No Custom Template seeds available' }];
    }
    const options = customTemplates.map((c) => ({
      value: c.Id,
      label: c.Title,
    }));
    return [{ value: 0, label: 'Select a Custom Template seed' }, ...options];
  }, [customTemplates]);

  const [fieldProps, meta, helpers] = useField<number>(
    'microk8s.customTemplateId'
  );
  useSetAvailableOption(
    customTemplateOptions,
    fieldProps.value,
    'microk8s.customTemplateId'
  );

  return (
    <FormControl
      label="Custom Template seed"
      tooltip="You can select a Custom Template to be auto installed when your environment is connected."
      inputId="kaas-customtemplate"
      errors={meta.error}
    >
      <Select
        name={fieldProps.name}
        id="kaas-customtemplate"
        data-cy="kaasCreateForm-customtemplateSelect"
        disabled={customTemplateOptions.length <= 1}
        options={customTemplateOptions}
        value={fieldProps.value}
        onChange={handleChange}
      />
    </FormControl>
  );

  function handleChange(e: ChangeEvent<HTMLSelectElement>) {
    const value = parseInt(e.target.value, 10);
    if (!Number.isNaN(value)) {
      helpers.setValue(value);
    }
  }
}
