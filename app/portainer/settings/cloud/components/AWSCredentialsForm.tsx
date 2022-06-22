import { Field, Form, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { Button } from '@@/buttons';

import { AccessKeyFormValues, KaasProvider, providerTitles } from '../types';

import { validationSchema } from './AWSCredentialsForm.validation';

type Props = {
  selectedProvider: KaasProvider;
  isEditing?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: AccessKeyFormValues) => void;
  credentialNames: string[];
  initialValues?: AccessKeyFormValues;
  placeholderText?: string;
};

const defaultInitialValues = {
  name: '',
  credentials: {
    accessKeyId: '',
    secretAccessKey: '',
  },
};

export function AWSCredentialsForm({
  selectedProvider,
  isEditing = false,
  isLoading,
  onSubmit,
  credentialNames,
  initialValues = defaultInitialValues,
  placeholderText = 'e.g. 5nq6uR3YQVhTNqRY2Q1lcft5rAp3Hq8S+8VUEDSW',
}: Props) {
  const router = useRouter();
  return (
    <Formik
      initialValues={initialValues}
      enableReinitialize
      validationSchema={() => validationSchema(credentialNames, isEditing)}
      onSubmit={(values) => onSubmit(values)}
      validateOnMount
    >
      {({ values, errors, handleSubmit, isValid, dirty }) => (
        <Form className="form-horizontal" onSubmit={handleSubmit} noValidate>
          {isEditing && (
            <FormControl inputId="provider" label="Provider">
              <Field
                as={Input}
                disabled
                name="provider"
                autoComplete="off"
                id="provider"
                value={providerTitles[selectedProvider]}
                data-cy="cloudSettings-provider"
              />
            </FormControl>
          )}
          <FormControl
            inputId="name"
            label="Credentials name"
            errors={errors.name}
          >
            <Field
              as={Input}
              name="name"
              id="name"
              value={values.name}
              placeholder={`e.g. ${providerTitles[selectedProvider]} Staging`}
              data-cy="cloudSettings-apiKeyName"
            />
          </FormControl>

          <>
            <FormControl
              inputId="access_key_id"
              label="Access key ID"
              errors={errors.credentials?.accessKeyId}
            >
              <Field
                as={Input}
                name="credentials.accessKeyId"
                autoComplete="off"
                id="access_key_id"
                value={values.credentials.accessKeyId}
                placeholder="e.g. AKIAUVTNNRIWDHKFJCXT"
                data-cy="cloudSettings-AccessKeyId"
              />
            </FormControl>
            <FormControl
              inputId="secret_access_key"
              label="Secret access key"
              errors={errors.credentials?.secretAccessKey}
            >
              <Field
                as={Input}
                name="credentials.secretAccessKey"
                autoComplete="off"
                id="secret_access_key"
                value={values.credentials.secretAccessKey}
                placeholder={placeholderText}
                data-cy="cloudSettings-SecretAccessKey"
              />
            </FormControl>
          </>

          <div className="form-group">
            <div className="col-sm-12 mt-3">
              <LoadingButton
                disabled={!isValid || !dirty}
                dataCy="createCredentials-saveButton"
                isLoading={isLoading}
                loadingText="Saving Credentials..."
              >
                {isEditing ? 'Update credentials' : 'Add credentials'}
              </LoadingButton>
              {isEditing && (
                <Button
                  color="default"
                  onClick={() =>
                    router.stateService.go('portainer.settings.cloud')
                  }
                >
                  Cancel
                </Button>
              )}
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
}
