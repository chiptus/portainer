import { Form, Formik } from 'formik';
import { useReducer } from 'react';

import { EdgeAsyncIntervalsForm } from '@/edge/components/EdgeAsyncIntervalsForm';
import { EdgeCheckinIntervalField } from '@/edge/components/EdgeCheckInIntervalField';
import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Switch } from '@/portainer/components/form-components/SwitchField/Switch';
import { Widget, WidgetBody, WidgetTitle } from '@/portainer/components/widget';
import { FormSection } from '@/portainer/components/form-components/FormSection';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { TextTip } from '@/portainer/components/Tip/TextTip';
import { notifySuccess } from '@/portainer/services/notifications';

import { useSettings, useUpdateSettingsMutation } from '../../queries';

import { FormValues } from './types';

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

export function DeploymentSyncOptions() {
  const settingsQuery = useSettings();
  const settingsMutation = useUpdateSettingsMutation();
  const [formKey, resetForm] = useReducer((state) => state + 1, 0);

  if (!settingsQuery.data) {
    return null;
  }

  const initialValues = {
    Edge: settingsQuery.data.Edge,
    EdgeAgentCheckinInterval: settingsQuery.data.EdgeAgentCheckinInterval,
  };

  return (
    <div className="row">
      <Widget>
        <WidgetTitle icon="fa-laptop" title="Deployment sync options" />
        <WidgetBody>
          <Formik<FormValues>
            initialValues={initialValues}
            onSubmit={handleSubmit}
            key={formKey}
          >
            {({ errors, setFieldValue, values, isValid, dirty }) => (
              <Form className="form-horizontal">
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

                <TextTip color="blue">
                  Using Async allows the ability to define different ping,
                  snapshot and command frequencies
                </TextTip>

                <FormSection title="Check-in Intervals">
                  {!values.Edge.AsyncMode ? (
                    <EdgeCheckinIntervalField
                      value={values.EdgeAgentCheckinInterval}
                      onChange={(value) =>
                        setFieldValue('EdgeAgentCheckinInterval', value)
                      }
                      isDefaultHidden
                      label="Edge agent default poll frequency"
                      tooltip="Interval used by default by each Edge agent to check in with the Portainer instance. Affects Edge environment management and Edge compute features."
                    />
                  ) : (
                    <EdgeAsyncIntervalsForm
                      values={values.Edge}
                      onChange={(value) => setFieldValue('Edge', value)}
                      isDefaultHidden
                      fieldSettings={asyncIntervalFieldSettings}
                    />
                  )}
                </FormSection>

                <FormSection title="Actions">
                  <div className="form-group mt-5">
                    <div className="col-sm-12">
                      <LoadingButton
                        disabled={!isValid || !dirty}
                        dataCy="settings-deploySyncOptionsButton"
                        isLoading={settingsMutation.isLoading}
                        loadingText="Saving settings..."
                      >
                        Save Settings
                      </LoadingButton>
                    </div>
                  </div>
                </FormSection>
              </Form>
            )}
          </Formik>
        </WidgetBody>
      </Widget>
    </div>
  );

  function handleSubmit(values: FormValues) {
    settingsMutation.mutate(
      {
        Edge: values.Edge,
        EdgeAgentCheckinInterval: values.EdgeAgentCheckinInterval,
      },
      {
        onSuccess() {
          notifySuccess('Settings updated successfully');
          resetForm();
        },
      }
    );
  }
}
