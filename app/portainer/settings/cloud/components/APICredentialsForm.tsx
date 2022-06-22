import { Field, Form, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { Button } from '@@/buttons';

import { APIFormValues, KaasProvider, providerTitles } from '../types';

import { validationSchema } from './APICredentialsForm.validation';

const exampleApiKeys: Partial<Record<KaasProvider, string>> = {
  civo: 'DgJ33kwIhnHumQFyc8ihGwWJql9cC8UJDiBhN8YImKqiX',
  linode: '92gsh9r9u5helgs4eibcuvlo403vm45hrmc6mzbslotnrqmkwc1ovqgmolcyq0wc',
  digitalocean:
    'dop_v1_n9rq7dkcbg3zb3bewtk9nnvmfkyfnr94d42n28lym22vhqu23rtkllsldygqm22v',
};

const defaultInitialValues = {
  name: '',
  credentials: {
    apiKey: '',
  },
};

type Props = {
  selectedProvider: KaasProvider;
  isEditing?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: APIFormValues) => void;
  credentialNames: string[];
  initialValues?: APIFormValues;
  placeholderText?: string;
};

export function APICredentialsForm({
  selectedProvider,
  isEditing = false,
  isLoading,
  onSubmit,
  credentialNames,
  initialValues = defaultInitialValues,
  placeholderText = `e.g. ${exampleApiKeys[selectedProvider]}`,
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
          <FormControl
            inputId="api_key"
            label="API key"
            errors={errors.credentials?.apiKey}
          >
            <Field
              as={Input}
              // see https://formik.org/docs/guides/arrays#nested-objects for the longer Name
              name="credentials.apiKey"
              autoComplete="off"
              id="api_key"
              value={values.credentials.apiKey}
              placeholder={placeholderText}
              data-cy="cloudSettings-ApiKey"
            />
          </FormControl>

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
