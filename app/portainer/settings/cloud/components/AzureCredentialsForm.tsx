import { Field, Form, Formik } from 'formik';

import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input } from '@/portainer/components/form-components/Input';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';

import { AzureFormValues, KaasProvider } from '../types';
import { getProviderTitle, isMeaningfulChange } from '../utils';

import { validationSchema } from './AzureCredentialsForm.validation';

type Props = {
  selectedProvider: KaasProvider;
  showProviderInput?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: AzureFormValues) => void;
  credentialNames: string[];
  initialValues?: AzureFormValues;
};

const defaultInitialValues = {
  name: '',
  credentials: {
    clientID: '',
    clientSecret: '',
    tenantID: '',
    subscriptionID: '',
  },
};

export function AzureCredentialsForm({
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
                value={getProviderTitle(selectedProvider)}
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
              placeholder={`e.g. ${getProviderTitle(selectedProvider)} Staging`}
              data-cy="cloudSettings-apiKeyName"
            />
          </FormControl>
          <FormControl
            inputId="Client_ID"
            label="Client ID"
            errors={errors.credentials?.clientID}
          >
            <Field
              as={Input}
              name="credentials.clientID"
              value={values.credentials.clientID}
              autoComplete="off"
              id="Client_ID"
              placeholder="e.g. b8fffb47-aed0-4723-bb81-3c2b15b275fb"
              data-cy="cloudSettings-ClientID"
            />
          </FormControl>
          <FormControl
            inputId="Client_Secret"
            label="Client Secret"
            errors={errors.credentials?.clientSecret}
          >
            <Field
              as={Input}
              name="credentials.clientSecret"
              value={values.credentials.clientSecret}
              autoComplete="off"
              id="Client_Secret"
              placeholder="e.g. WWE8Q~0J4GdItGa2UZiZU6pFYewu7c4cvmCSPbZF"
              data-cy="cloudSettings-ClientSecret"
            />
          </FormControl>
          <FormControl
            inputId="Tenant_ID"
            label="Tenant ID"
            errors={errors.credentials?.tenantID}
          >
            <Field
              as={Input}
              name="credentials.tenantID"
              value={values.credentials.tenantID}
              autoComplete="off"
              id="Tenant_ID"
              placeholder="e.g. bea09a7f-8bc0-4e95-b130-762078e972ef"
              data-cy="cloudSettings-TenantID"
            />
          </FormControl>
          <FormControl
            inputId="Subscription_ID"
            label="Subscription ID"
            errors={errors.credentials?.subscriptionID}
          >
            <Field
              as={Input}
              name="credentials.subscriptionID"
              value={values.credentials.subscriptionID}
              autoComplete="off"
              id="Subscription_ID"
              placeholder="e.g. c57c21b6-0e0d-448d-aa44-d33a1b9ab5e4"
              data-cy="cloudSettings-SubscriptionID"
            />
          </FormControl>

          <div className="form-group">
            <div className="col-sm-12 mt-20">
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
