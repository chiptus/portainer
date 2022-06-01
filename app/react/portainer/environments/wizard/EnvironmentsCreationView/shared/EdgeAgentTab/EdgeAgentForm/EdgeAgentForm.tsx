import { Formik, Form } from 'formik';

import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { Environment } from '@/portainer/environments/types';
import { useCreateEdgeAgentEnvironmentMutation } from '@/portainer/environments/queries/useCreateEnvironmentMutation';
import { baseHref } from '@/portainer/helpers/pathHelper';
import { FormSection } from '@/portainer/components/form-components/FormSection';
import { EdgeCheckinIntervalField } from '@/edge/components/EdgeCheckInIntervalField';
import {
  EdgeAsyncIntervalsForm,
  EDGE_ASYNC_INTERVAL_USE_DEFAULT,
} from '@/edge/components/EdgeAsyncIntervalsForm';
import { useSettings } from '@/portainer/settings/queries';

import { MoreSettingsSection } from '../../MoreSettingsSection';

import { EdgeAgentFieldset } from './EdgeAgentFieldset';
import { validationSchema } from './EdgeAgentForm.validation';
import { FormValues } from './types';

interface Props {
  onCreate(environment: Environment): void;
  readonly: boolean;
}

const initialValues = buildInitialValues();

export function EdgeAgentForm({ onCreate, readonly }: Props) {
  const edgeSettingsQuery = useSettings((settings) => settings.Edge);

  const createMutation = useCreateEdgeAgentEnvironmentMutation();

  if (!edgeSettingsQuery.data) {
    return null;
  }

  const edgeSettings = edgeSettingsQuery.data;

  return (
    <Formik<FormValues>
      initialValues={initialValues}
      onSubmit={handleSubmit}
      validateOnMount
      validationSchema={validationSchema}
    >
      {({ isValid, setFieldValue, values }) => (
        <Form>
          <EdgeAgentFieldset readonly={readonly} />

          <MoreSettingsSection>
            <FormSection title="Check-in Intervals">
              {edgeSettings.AsyncMode ? (
                <EdgeAsyncIntervalsForm
                  values={values.edge}
                  readonly={readonly}
                  onChange={(values) => setFieldValue('edge', values)}
                />
              ) : (
                <EdgeCheckinIntervalField
                  readonly={readonly}
                  onChange={(value) => setFieldValue('pollFrequency', value)}
                  value={values.pollFrequency}
                />
              )}
            </FormSection>
          </MoreSettingsSection>

          {!readonly && (
            <div className="row">
              <div className="col-sm-12">
                <LoadingButton
                  isLoading={createMutation.isLoading}
                  loadingText="Creating environment..."
                  disabled={!isValid}
                >
                  <i className="fa fa-plug space-right" />
                  Create
                </LoadingButton>
              </div>
            </div>
          )}
        </Form>
      )}
    </Formik>
  );

  function handleSubmit(values: typeof initialValues) {
    createMutation.mutate(values, {
      onSuccess(environment) {
        onCreate(environment);
      },
    });
  }
}

export function buildInitialValues(): FormValues {
  return {
    name: '',
    portainerUrl: defaultPortainerUrl(),
    pollFrequency: 0,
    meta: {
      groupId: 1,
      tagIds: [],
    },
    edge: {
      CommandInterval: EDGE_ASYNC_INTERVAL_USE_DEFAULT,
      PingInterval: EDGE_ASYNC_INTERVAL_USE_DEFAULT,
      SnapshotInterval: EDGE_ASYNC_INTERVAL_USE_DEFAULT,
    },
  };

  function defaultPortainerUrl() {
    const baseHREF = baseHref();
    return window.location.origin + (baseHREF !== '/' ? baseHREF : '');
  }
}
