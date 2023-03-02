import { Form, Formik } from 'formik';
import { useReducer } from 'react';
import { Laptop } from 'lucide-react';

import { EdgeCheckinIntervalField } from '@/react/edge/components/EdgeCheckInIntervalField';
import { EdgeAsyncIntervalsForm } from '@/react/edge/components/EdgeAsyncIntervalsForm';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { MTLSOptions } from '@/react/edge/components/MTLSOptions';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';

import { Widget, WidgetBody, WidgetTitle } from '@@/Widget';
import { FormSection } from '@@/form-components/FormSection';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { TextTip } from '@@/Tip/TextTip';

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

  const initialValues: FormValues = {
    Edge: {
      CommandInterval: settingsQuery.data.Edge.CommandInterval,
      PingInterval: settingsQuery.data.Edge.PingInterval,
      SnapshotInterval: settingsQuery.data.Edge.SnapshotInterval,
      MTLS: settingsQuery.data.Edge.MTLS,
    },
    EdgeAgentCheckinInterval: settingsQuery.data.EdgeAgentCheckinInterval,
  };

  return (
    <div className="row">
      <Widget>
        <WidgetTitle icon={Laptop} title="Deployment sync options" />
        <WidgetBody>
          <Formik<FormValues>
            initialValues={initialValues}
            onSubmit={handleSubmit}
            key={formKey}
          >
            {({ setFieldValue, values, isValid, dirty }) => (
              <Form className="form-horizontal">
                <TextTip color="blue">
                  Default values set here will be available to choose as an
                  option for edge environment creation
                </TextTip>

                <FormSection title="Check-in Intervals">
                  <EdgeCheckinIntervalField
                    value={values.EdgeAgentCheckinInterval}
                    onChange={(value) =>
                      setFieldValue('EdgeAgentCheckinInterval', value)
                    }
                    isDefaultHidden
                    label="Edge agent default poll frequency"
                    tooltip="Interval used by default by each Edge agent to check in with the Portainer instance. Affects Edge environment management and Edge compute features."
                  />
                </FormSection>

                {isBE && (
                  <FormSection title="Async Check-in Intervals">
                    <EdgeAsyncIntervalsForm
                      values={values.Edge}
                      onChange={(value) => setFieldValue('Edge', value)}
                      isDefaultHidden
                      fieldSettings={asyncIntervalFieldSettings}
                    />
                  </FormSection>
                )}

                <FormSection title="mTLS Certificate">
                  <MTLSOptions
                    values={values.Edge.MTLS}
                    onChange={(value) => setFieldValue('Edge.MTLS', value)}
                  />
                </FormSection>

                <div>
                  <div className="form-group mt-5">
                    <div className="col-sm-12">
                      <LoadingButton
                        disabled={!isValid || !dirty}
                        className="!ml-0"
                        data-cy="settings-deploySyncOptionsButton"
                        isLoading={settingsMutation.isLoading}
                        loadingText="Saving settings..."
                      >
                        Save settings
                      </LoadingButton>
                    </div>
                  </div>
                </div>
              </Form>
            )}
          </Formik>
        </WidgetBody>
      </Widget>
    </div>
  );

  async function handleSubmit(values: FormValues) {
    let mtlsValues = {
      UseSeparateCert: false,
      CaCert: '',
      Cert: '',
      Key: '',
    };

    if (values.Edge.MTLS.UseSeparateCert) {
      const caCert = values.Edge.MTLS.CaCertFile
        ? values.Edge.MTLS.CaCertFile.text()
        : '';
      const cert = values.Edge.MTLS.CertFile
        ? values.Edge.MTLS.CertFile.text()
        : '';
      const key = values.Edge.MTLS.KeyFile
        ? values.Edge.MTLS.KeyFile.text()
        : '';

      mtlsValues = {
        ...values.Edge.MTLS,
        CaCert: await caCert,
        Cert: await cert,
        Key: await key,
      };
    }

    settingsMutation.mutate(
      {
        Edge: {
          ...values.Edge,
          MTLS: mtlsValues,
        },
        EdgeAgentCheckinInterval: values.EdgeAgentCheckinInterval,
      },
      {
        onSuccess() {
          notifySuccess('Success', 'Settings updated successfully');
          resetForm();
        },
        onError(error) {
          notifyError('Failure', error as Error, 'Unable to update settings');
        },
      }
    );
  }
}
