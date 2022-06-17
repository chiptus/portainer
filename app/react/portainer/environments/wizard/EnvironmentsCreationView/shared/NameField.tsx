import { Field, useField } from 'formik';
import { string } from 'yup';

import { getEndpoints } from '@/portainer/environments/environment.service';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';

interface Props {
  readonly?: boolean;
  tooltip?: string;
  placeholder?: string;
}

export function NameField({
  readonly,
  tooltip,
  placeholder = 'e.g. docker-prod01 / kubernetes-cluster01',
}: Props) {
  const [, meta] = useField('name');

  const id = 'name-input';

  return (
    <FormControl
      label="Name"
      required
      errors={meta.error}
      inputId={id}
      tooltip={tooltip}
    >
      <Field
        id={id}
        name="name"
        as={Input}
        data-cy="endpointCreate-nameInput"
        placeholder={placeholder}
        readOnly={readonly}
      />
    </FormControl>
  );
}

async function isNameUnique(name?: string) {
  if (!name) {
    return true;
  }

  try {
    const result = await getEndpoints(0, 1, { name });
    if (result.totalCount > 0) {
      return false;
    }
  } catch (e) {
    // if backend fails to respond, assume name is unique, name validation happens also in the backend
  }
  return true;
}

export function nameValidation() {
  return string()
    .required('Name is required')
    .test(
      'unique-name',
      'Name should be unique',
      async (name) => (await isNameUnique(name)) || false
    );
}
