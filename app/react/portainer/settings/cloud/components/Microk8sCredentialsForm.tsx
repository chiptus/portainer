import { Field, Form, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { Button } from '@@/buttons';

import {
  KaasProvider,
  providerTitles,
  UsernamePasswordFormValues,
} from '../types';

import { validationSchema } from './Microk8sCredentialsForm.validation';

type Props = {
  selectedProvider: KaasProvider;
  isEditing?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: UsernamePasswordFormValues) => void;
  credentialNames: string[];
  initialValues?: UsernamePasswordFormValues;
  placeholderText?: string;
};

const defaultInitialValues = {
  name: '',
  credentials: {
    username: '',
    password: '',
  },
};

export function Microk8sCredentialsForm({
  selectedProvider,
  isEditing = false,
  isLoading,
  onSubmit,
  credentialNames,
  initialValues = defaultInitialValues,
  placeholderText = '********',
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
              inputId="ssh_username"
              label="SSH username"
              errors={errors.credentials?.username}
            >
              <Field
                as={Input}
                name="credentials.username"
                autoComplete="off"
                id="ssh_username"
                value={values.credentials.username}
                placeholder="e.g. ssh-builder"
                data-cy="cloudSettings-SshUsername"
              />
            </FormControl>
            <FormControl
              inputId="ssh_password"
              label="SSH password"
              errors={errors.credentials?.password}
            >
              <Field
                as={Input}
                name="credentials.password"
                autoComplete="off"
                type="password"
                id="ssh_password"
                value={values.credentials.password}
                placeholder={placeholderText}
                data-cy="cloudSettings-SshPassword"
              />
            </FormControl>
          </>

          <div className="form-group">
            <div className="col-sm-12 mt-3">
              <LoadingButton
                disabled={!isValid || !dirty}
                data-cy="createCredentials-saveButton"
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
