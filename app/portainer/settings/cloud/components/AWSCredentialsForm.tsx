import { Field, Form, Formik } from 'formik';

import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input } from '@/portainer/components/form-components/Input';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';

import { AccessKeyFormValues, KaasProvider, providerTitles } from '../types';
import { isMeaningfulChange } from '../utils';

import { validationSchema } from './AWSCredentialsForm.validation';

type Props = {
  selectedProvider: KaasProvider;
  showProviderInput?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: AccessKeyFormValues) => void;
  credentialNames: string[];
  initialValues?: AccessKeyFormValues;
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
  showProviderInput = false,
  isLoading,
  onSubmit,
  credentialNames,
  initialValues = defaultInitialValues,
}: Props) {
  return (
    <Formik
      initialValues={initialValues}
      enableReinitialize
      validationSchema={() => validationSchema(credentialNames)}
      onSubmit={(values) => onSubmit(values)}
      validateOnMount
    >
      {({ values, errors, handleSubmit, isValid, dirty }) => (
        <Form className="form-horizontal" onSubmit={handleSubmit} noValidate>
          {showProviderInput && (
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
          <FormControl inputId="Name" label="Name" errors={errors.name}>
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
              label="Access Key Id"
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
              label="Secret Access Key"
              errors={errors.credentials?.secretAccessKey}
            >
              <Field
                as={Input}
                name="credentials.secretAccessKey"
                autoComplete="off"
                id="secret_access_key"
                value={values.credentials.secretAccessKey}
                placeholder="e.g. 5nq6uR3YQVhTNqRY2Q1lcft5rAp3Hq8S+8VUEDSW"
                data-cy="cloudSettings-SecretAccessKey"
              />
            </FormControl>
          </>

          <div className="form-group">
            <div className="col-sm-12 mt-3">
              <LoadingButton
                disabled={
                  !isValid ||
                  !dirty ||
                  !isMeaningfulChange(values, initialValues)
                }
                dataCy="createCredentials-saveButton"
                isLoading={isLoading}
                loadingText="Saving Credentials..."
              >
                Save
              </LoadingButton>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
}
