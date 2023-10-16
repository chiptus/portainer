import { Form as FormikForm, Formik } from 'formik';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import { notifySuccess } from '@/portainer/services/notifications';
import { withLimitToBE } from '@/react/hooks/useLimitToBE';
import {
  HttpsWarning,
  isHttps,
} from '@/react/edge/edge-configurations/common/HttpsWarning';

import { PageHeader } from '@@/PageHeader';
import { Widget } from '@@/Widget';
import { FormActions } from '@@/form-components/FormActions';

import { useUpdateMutation } from '../queries/update/update';
import {
  FormValues,
  FormValuesEdgeConfigurationMatchingRule,
  FormValuesEdgeConfigurationType,
} from '../common/types';
import { EdgeGroupsField } from '../common/EdgeGroupsField';
import { EdgeConfigurationTypeSelectorField } from '../common/EdgeConfigurationTypeSelectorField';
import { ConfigurationFieldset } from '../common/ConfigurationFieldset';
import { InputField } from '../common/InputField';
import { useEdgeConfiguration } from '../queries/item/item';
import { EdgeConfigurationCategory, EdgeConfigurationType } from '../types';

import { validation } from './validation';

export default withLimitToBE(ItemView);

function buildFormValuesType(
  type: EdgeConfigurationType
): Pick<FormValues, 'type' | 'matchingRule'> {
  if (type === EdgeConfigurationType.EdgeConfigTypeSpecificFile) {
    return {
      type: FormValuesEdgeConfigurationType.DeviceSpecific,
      matchingRule: FormValuesEdgeConfigurationMatchingRule.MatchFile,
    };
  }
  if (type === EdgeConfigurationType.EdgeConfigTypeSpecificFolder) {
    return {
      type: FormValuesEdgeConfigurationType.DeviceSpecific,
      matchingRule: FormValuesEdgeConfigurationMatchingRule.MatchFolder,
    };
  }
  return {
    type: FormValuesEdgeConfigurationType.General,
  };
}

function ItemView() {
  const router = useRouter();
  const {
    params: { id },
  } = useCurrentStateAndParams();
  const updateMutation = useUpdateMutation();
  const edgeConfigQuery = useEdgeConfiguration(id);

  if (!edgeConfigQuery.data) {
    return null;
  }

  const { name, edgeGroupIDs, type, category, baseDir } = edgeConfigQuery.data;

  const showHttpsWarning =
    !isHttps() && category === EdgeConfigurationCategory.Secret;

  const initialValues: FormValues = {
    name,
    groupIds: edgeGroupIDs,
    ...buildFormValuesType(type),
    category: EdgeConfigurationCategory.Configuration,
    directory: baseDir,
    file: { name: '' },
  };

  return (
    <>
      <PageHeader
        title={`Edit edge ${category}`}
        breadcrumbs={[
          { label: 'Edge configurations', link: 'edge.configurations' },
          { label: `Edit edge ${category}` },
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
                    <InputField
                      fieldName="name"
                      label="Name"
                      required
                      disabled
                    />
                    <EdgeGroupsField />
                    <InputField
                      fieldName="directory"
                      label="Directory"
                      placeholder="/etc/edge"
                      tooltip={`A designated folder on each edge node for storing read-only ${category} files.`}
                      required
                      disabled
                    />

                    <div className="mt-2">
                      <EdgeConfigurationTypeSelectorField category={category} />
                    </div>

                    <ConfigurationFieldset category={category} />

                    <FormActions
                      submitLabel={`Update ${category} and push`}
                      loadingText="Updating..."
                      isValid={isValid && !showHttpsWarning}
                      isLoading={updateMutation.isLoading}
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
    updateMutation.mutate(
      { id, values },
      {
        onSuccess() {
          notifySuccess('Success', `Successfully updated edge ${category}`);
          router.stateService.reload();
        },
      }
    );
  }
}
