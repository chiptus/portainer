import { Formik, Form } from 'formik';

import { Switch } from '@/portainer/components/form-components/SwitchField/Switch';
import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Widget, WidgetBody, WidgetTitle } from '@/portainer/components/widget';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { TextTip } from '@/portainer/components/Tip/TextTip';
import { EdgeAsyncIntervalsForm } from '@/edge/components/EdgeAsyncIntervalsForm';
import { EdgeCheckinIntervalField } from '@/edge/components/EdgeCheckInIntervalField';
import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';

import { validationSchema } from './EdgeComputeSettings.validation';
import { FormValues } from './types';

interface Props {
  settings?: FormValues;
  onSubmit(values: FormValues): void;
}

const asyncIntervalFieldSettings = {
  ping: {
    label: 'Edge agent default ping frequency',
    tooltip:
      'Interval used by default by each Edge agent to ping the Portainer instance. Affects Edge environment management and Edge compute features.',
  },
  snapshot: {
    label: 'Edge agent default snapshot frequency',
    tooltip:
      'Interval used by default by each Edge agent to snapshot the agent state.',
  },
  command: {
    label: 'Edge agent default command frequency',
    tooltip: 'Interval used by default by each Edge agent to execute commands.',
  },
};

export function EdgeComputeSettings({ settings, onSubmit }: Props) {
  if (!settings) {
    return null;
  }

  return (
    <div className="row">
      <Widget>
        <WidgetTitle icon="fa-laptop" title="Edge Compute settings" />
        <WidgetBody>
          <Formik
            initialValues={settings}
            enableReinitialize
            validationSchema={() => validationSchema()}
            onSubmit={onSubmit}
            validateOnMount
          >
            {({
              values,
              errors,
              handleSubmit,
              setFieldValue,
              isSubmitting,
              isValid,
              dirty,
            }) => (
              <Form
                className="form-horizontal"
                onSubmit={handleSubmit}
                noValidate
              >
                <FormControl
                  inputId="edge_enable"
                  label="Enable Edge Compute features"
                  size="medium"
                  errors={errors.EnableEdgeComputeFeatures}
                >
                  <Switch
                    id="edge_enable"
                    name="edge_enable"
                    className="space-right"
                    checked={values.EnableEdgeComputeFeatures}
                    onChange={(e) =>
                      setFieldValue('EnableEdgeComputeFeatures', e)
                    }
                  />
                </FormControl>

                <TextTip color="blue">
                  When enabled, this will enable Portainer to execute Edge
                  Device features.
                </TextTip>

                <FormControl
                  inputId="edge_enforce_id"
                  label="Enforce use of Portainer generated Edge ID"
                  size="medium"
                  tooltip="This setting only applies to manually created environments."
                  errors={errors.EnforceEdgeID}
                >
                  <Switch
                    id="edge_enforce_id"
                    name="edge_enforce_id"
                    className="space-right"
                    checked={values.EnforceEdgeID}
                    onChange={(e) =>
                      setFieldValue('EnforceEdgeID', e.valueOf())
                    }
                  />
                </FormControl>

                {process.env.PORTAINER_EDITION === 'BE' && (
                  <FormControl
                    inputId="edge_async_mode"
                    label="Use Async mode by default"
                    size="medium"
                    errors={errors?.Edge?.AsyncMode}
                  >
                    <Switch
                      id="edge_async_mode"
                      name="edge_async_mode"
                      className="space-right"
                      checked={values.Edge.AsyncMode}
                      onChange={(e) =>
                        setFieldValue('Edge.AsyncMode', e.valueOf())
                      }
                    />
                  </FormControl>
                )}

                <FormSectionTitle>Check-in Intervals</FormSectionTitle>

                <EdgeCheckinIntervalField
                  value={values.EdgeAgentCheckinInterval}
                  onChange={(value) =>
                    setFieldValue('EdgeAgentCheckinInterval', value)
                  }
                  isDefaultHidden
                  label="Edge agent default poll frequency"
                  tooltip="Interval used by default by each Edge agent to check in with the Portainer instance. Affects Edge environment management and Edge compute features."
                />

                <EdgeAsyncIntervalsForm
                  values={values.Edge}
                  onChange={(value) => setFieldValue('Edge', value)}
                  isDefaultHidden
                  fieldSettings={asyncIntervalFieldSettings}
                />

                <div className="form-group mt-5">
                  <div className="col-sm-12">
                    <LoadingButton
                      disabled={!isValid || !dirty}
                      dataCy="settings-edgeComputeButton"
                      isLoading={isSubmitting}
                      loadingText="Saving settings..."
                    >
                      Save Settings
                    </LoadingButton>
                  </div>
                </div>
              </Form>
            )}
          </Formik>
        </WidgetBody>
      </Widget>
    </div>
  );
}
