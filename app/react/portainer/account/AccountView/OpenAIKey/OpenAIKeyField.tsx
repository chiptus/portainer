import { useField } from 'formik';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';

const fieldKey = 'OpenAIApiKey';
const id = 'user_openai_key';
export function OpenAIKeyField() {
  const [inputProps, meta, helpers] = useField<string>(fieldKey);

  return (
    <FormControl
      inputId={id}
      label="OpenAI key"
      size="xsmall"
      errors={meta.error}
    >
      <Input
        name={id}
        id={id}
        className="space-right"
        value={inputProps.value}
        onChange={(e) => helpers.setValue(e.target.value)}
      />
    </FormControl>
  );
}
