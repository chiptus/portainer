import { Field, useField } from 'formik';
import { ComponentProps } from 'react';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';

// don't allow to manually set id and name as the component controls both fields
type InputProps = Omit<ComponentProps<typeof Input>, 'id' | 'name'>;

type FormControlProps = {
  label: string;
  fieldName: string;
  tooltip?: string;
};

type Props = FormControlProps & InputProps;

export function InputField({
  label,
  fieldName,
  tooltip,
  required,
  ...props
}: Props) {
  const [{ name }, { error }] = useField(fieldName);

  return (
    <FormControl
      label={label}
      required={required}
      inputId={`${fieldName}-input`}
      errors={error}
      tooltip={tooltip}
    >
      <Field
        as={Input}
        name={name}
        id={`${fieldName}-input`}
        // eslint-disable-next-line react/jsx-props-no-spreading
        {...props}
      />
    </FormControl>
  );
}
