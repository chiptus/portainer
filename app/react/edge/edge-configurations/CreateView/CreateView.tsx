import { Form as FormikForm, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { notifySuccess } from '@/portainer/services/notifications';
import { withLimitToBE } from '@/react/hooks/useLimitToBE';
import { useCategory } from '@/react/edge/edge-configurations/CreateView/useCategory';
import {
  HttpsWarning,
  isHttps,
} from '@/react/edge/edge-configurations/common/HttpsWarning';
import { EdgeConfigurationCategory } from '@/react/edge/edge-configurations/types';

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
  category: EdgeConfigurationCategory.Configuration,
  directory: '',
  file: { name: '' },
};

function CreateView() {
  const createMutation = useCreateMutation();
  const router = useRouter();

  const [category] = useCategory();

  initialValues.category = category;

  const showHttpsWarning =
    !isHttps() && category === EdgeConfigurationCategory.Secret;

  return (
    <>
      <PageHeader
        title={`Create edge ${category}`}
        breadcrumbs={[
          { label: `Edge ${category}s`, link: 'edge.configurations' },
          { label: `Create edge ${category}` },
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
                      tooltip={`A designated folder on each edge node for storing read-only ${category} files.`}
                      required
                    />

                    <div className="mt-2">
                      <EdgeConfigurationTypeSelectorField category={category} />
                    </div>

                    <ConfigurationFieldset category={category} />

                    <FormActions
                      submitLabel={`Create ${category} and push`}
                      loadingText="Creating..."
                      isValid={isValid && !showHttpsWarning}
                      isLoading={createMutation.isLoading}
                    />

                    {showHttpsWarning && <HttpsWarning />}
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
        notifySuccess('Success', `Successfully created edge ${category}`);
        router.stateService.go('^');
      },
    });
  }
}
