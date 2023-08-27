import { Formik, Form as FormikForm } from 'formik';
import { useRouter } from '@uirouter/react';

import { notifySuccess } from '@/portainer/services/notifications';
import { withLimitToBE } from '@/react/hooks/useLimitToBE';

import { PageHeader } from '@@/PageHeader';
import { Widget } from '@@/Widget';
import { FormActions } from '@@/form-components/FormActions';

import { useCreateMutation } from '../queries/create/create';
import { FormValues, FormValuesEdgeConfigurationType } from '../common/types';
import { EdgeGroupsField } from '../common/EdgeGroupsField';
import { EdgeConfigurationTypeSelectorField } from '../common/EdgeConfigurationTypeSelectorField';
import { ConfigurationFieldset } from '../common/ConfigurationFieldset';
import { InputField } from '../common/InputField';

import { validation } from './validation';

export default withLimitToBE(CreateView);

const initialValues: FormValues = {
  name: '',
  groupIds: [],
  type: FormValuesEdgeConfigurationType.General,
  directory: '',
  file: { name: '' },
};

function CreateView() {
  const createMutation = useCreateMutation();
  const router = useRouter();

  return (
    <>
      <PageHeader
        title="Create edge configuration"
        breadcrumbs={[
          { label: 'Edge configurations', link: 'edge.configurations' },
          { label: 'Create edge configuration' },
        ]}
        reload
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <Widget.Body>
              <Formik
                initialValues={initialValues}
                onSubmit={handleSubmit}
                validateOnMount
                validationSchema={() => validation()}
              >
                {({ isValid }) => (
                  <FormikForm className="form-horizontal">
                    <InputField fieldName="name" label="Name" required />
                    <EdgeGroupsField />
                    <InputField
                      fieldName="directory"
                      label="Directory"
                      placeholder="/etc/edge"
                      tooltip="A designated folder on each edge node for storing read-only configuration files."
                      required
                    />

                    <div className="mt-2">
                      <EdgeConfigurationTypeSelectorField />
                    </div>

                    <ConfigurationFieldset />

                    <FormActions
                      submitLabel="Create configuration and push"
                      loadingText="Creating..."
                      isValid={isValid}
                      isLoading={createMutation.isLoading}
                    />
                  </FormikForm>
                )}
              </Formik>
            </Widget.Body>
          </Widget>
        </div>
      </div>
    </>
  );

  function handleSubmit(values: FormValues) {
    createMutation.mutate(values, {
      onSuccess() {
        notifySuccess('Success', 'Successfully created edge configuration');
        router.stateService.go('^');
      },
    });
  }
}
