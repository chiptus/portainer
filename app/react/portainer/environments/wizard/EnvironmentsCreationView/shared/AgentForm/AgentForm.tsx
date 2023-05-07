import { Form, Formik } from 'formik';
import { useReducer } from 'react';
import { Plug2 } from 'lucide-react';

import { useCreateAgentEnvironmentMutation } from '@/react/portainer/environments/queries/useCreateEnvironmentMutation';
import { notifySuccess } from '@/portainer/services/notifications';
import { Environment } from '@/react/portainer/environments/types';
import { CreateAgentEnvironmentValues } from '@/react/portainer/environments/environment.service/create';
import { CustomTemplate } from '@/react/portainer/custom-templates/types';

import { LoadingButton } from '@@/buttons/LoadingButton';

import { NameField } from '../NameField';
import { MoreSettingsSection } from '../MoreSettingsSection';
import { CustomTemplateSelector } from '../CustomTemplateSelector';

import { EnvironmentUrlField } from './EnvironmentUrlField';
import { useValidation } from './AgentForm.validation';

interface Props {
  onCreate(environment: Environment): void;
  customTemplates?: CustomTemplate[];
}

const initialValues: CreateAgentEnvironmentValues = {
  environmentUrl: '',
  name: '',
  meta: {
    groupId: 1,
    tagIds: [],
    customTemplateId: 0,
    variables: {},
  },
};

export function AgentForm({ onCreate, customTemplates }: Props) {
  const [formKey, clearForm] = useReducer((state) => state + 1, 0);

  const mutation = useCreateAgentEnvironmentMutation();
  const validation = useValidation();

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={handleSubmit}
      validationSchema={validation}
      validateOnMount
      key={formKey}
    >
      {({ isValid, dirty }) => (
        <Form>
          <NameField />
          <EnvironmentUrlField />

          <MoreSettingsSection>
            {customTemplates && (
              <CustomTemplateSelector customTemplates={customTemplates} />
            )}
          </MoreSettingsSection>

          <div className="form-group">
            <div className="col-sm-12">
              <LoadingButton
                className="wizard-connect-button vertical-center"
                loadingText="Connecting environment..."
                isLoading={mutation.isLoading}
                disabled={!dirty || !isValid}
                icon={Plug2}
              >
                Connect
              </LoadingButton>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );

  function handleSubmit(values: CreateAgentEnvironmentValues) {
    mutation.mutate(values, {
      onSuccess(environment) {
        notifySuccess('Environment created', environment.Name);
        clearForm();
        onCreate(environment);
      },
    });
  }
}
