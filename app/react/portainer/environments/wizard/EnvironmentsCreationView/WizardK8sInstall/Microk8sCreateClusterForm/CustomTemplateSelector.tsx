import { useField } from 'formik';
import { ChangeEvent, useMemo } from 'react';

import { CustomTemplate } from '@/react/portainer/custom-templates/types';

import { FormControl } from '@@/form-components/FormControl';
import { Select, Option } from '@@/form-components/Input/Select';

import { useSetAvailableOption } from '../../WizardKaaS/useSetAvailableOption';

interface Props {
  customTemplates: CustomTemplate[];
}

export function CustomTemplateSelector({ customTemplates }: Props) {
  const customTemplateOptions: Option<number>[] = useMemo(() => {
    if (customTemplates.length === 0) {
      return [{ value: 0, label: 'No Custom Templates available' }];
    }
    const options = customTemplates.map((c) => ({
      value: c.Id,
      label: c.Title,
    }));
    return [{ value: 0, label: 'Select a Custom Template' }, ...options];
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
      label="Custom Template"
      tooltip={
        <>
          <div>
            You can select a Custom Template to be auto installed when your
            environment is connected.
          </div>
          {customTemplateOptions.length <= 1 && (
            <div className="mt-2">
              You don&apos;t currently have any, but you can create them via the
              Custom Templates menu option, once you have at least one
              Kubernetes environment set up.
            </div>
          )}
        </>
      }
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
