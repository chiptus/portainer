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
      tooltip={
        <p>
          Set an OpenAI API key, copied from your{' '}
          <a
            href="https://platform.openai.com/account/api-keys"
            target="_blank"
            rel="noreferrer"
          >
            OpenAI Platform Account
          </a>
          .
        </p>
      }
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
