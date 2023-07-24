import { useField } from 'formik';

import { FormControl } from '@@/form-components/FormControl';
import { Switch } from '@@/form-components/SwitchField/Switch';

const fieldKey = 'OpenAIIntegration';

export function EnableOpenAIIntegrationSwitch() {
  const [inputProps, meta, helpers] = useField<boolean>(fieldKey);

  return (
    <FormControl
      inputId="experimental_openAI"
      label="Enable OpenAI integration"
      tooltip="Users can set their OpenAI API key via their User settings page (accessed via the My Account option from the top-right of the screen)."
      size="medium"
      errors={meta.error}
    >
      <Switch
        id="experimental_openAI"
        name={fieldKey}
        className="space-right"
        checked={inputProps.value}
        onChange={handleChange}
      />
    </FormControl>
  );

  function handleChange(enable: boolean) {
    helpers.setValue(enable);
  }
}
