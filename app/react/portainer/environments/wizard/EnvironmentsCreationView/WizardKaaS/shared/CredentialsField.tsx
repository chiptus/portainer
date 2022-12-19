import { useField } from 'formik';
import { ChangeEvent } from 'react';

import { Credential } from '@/react/portainer/settings/cloud/types';

import { FormControl } from '@@/form-components/FormControl';
import { Select, Option } from '@@/form-components/Input/Select';

import { useSetAvailableOption } from '../useSetAvailableOption';

interface Props {
  credentials: Credential[];
}

export function CredentialsField({ credentials }: Props) {
  const credentialOptions: Option<number>[] = credentials.map((c) => ({
    value: c.id,
    label: c.name,
  }));

  const [fieldProps, meta, helpers] = useField<number>('credentialId');
  useSetAvailableOption(credentialOptions, fieldProps.value, 'credentialId');

  return (
    <FormControl
      label="Credentials"
      tooltip="Credentials to create your cluster."
      inputId="kaas-credential"
      errors={meta.error}
    >
      <Select
        name={fieldProps.name}
        id="kaas-credential"
        data-cy="kaasCreateForm-credentialSelect"
        disabled={credentialOptions.length <= 1}
        options={credentialOptions}
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
