import { Field, Form, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { Button } from '@@/buttons';

import {
  CredentialType,
  credentialTitles,
  SSHCredentialFormValues,
} from '../../types';

import { validationSchema } from './SSHCredentialsForm.validation';
import SSHCredentialsPrivateKeyForm from './SSHCredentialsPrivateKeyForm';

type Props = {
  selectedProvider: CredentialType;
  isEditing?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: SSHCredentialFormValues) => void;
  credentialNames: string[];
  initialValues?: SSHCredentialFormValues;
  hasSSHKey?: boolean;
  hasPassphrase?: boolean;
};

const defaultInitialValues = {
  name: '',
  credentials: {
    username: '',
    password: '',
    privateKey: '',
    passphrase: '',
  },
};

export function SSHCredentialsForm({
  selectedProvider,
  isEditing = false,
  isLoading,
  onSubmit,
  credentialNames,
  initialValues = defaultInitialValues,
  hasSSHKey,
  hasPassphrase,
}: Props) {
  const router = useRouter();
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
          {isEditing && (
            <FormControl inputId="provider" label="Provider" size="medium">
              <Field
                as={Input}
                disabled
                name="provider"
                autoComplete="off"
                id="provider"
                value={credentialTitles[selectedProvider]}
                data-cy="cloudSettings-provider"
              />
            </FormControl>
          )}
          <FormControl
            inputId="name"
            label="Credentials name"
            size="medium"
            errors={errors.name}
            required
          >
            <Field
              as={Input}
              name="name"
              id="name"
              maxLength={200}
              value={values.name}
              placeholder={`e.g. ${credentialTitles[selectedProvider]} Staging`}
              data-cy="cloudSettings-apiKeyName"
            />
          </FormControl>

          <>
            <FormControl
              inputId="ssh_username"
              label="SSH username"
              size="medium"
              tooltip="The username on the host machine that you will connect to. This user must have sudo privileges."
              errors={errors.credentials?.username}
              required
            >
              <Field
                as={Input}
                name="credentials.username"
                autoComplete="off"
                id="ssh_username"
                maxLength={200}
                value={values.credentials.username}
                placeholder="e.g. ssh_builder"
                data-cy="cloudSettings-SshUsername"
              />
            </FormControl>
            <FormControl
              inputId="ssh_password"
              label="SSH password"
              size="medium"
              tooltip="The password used to run privileged commands on the host machine. It is also used for SSH password authentication, if no SSH key is used."
              errors={errors.credentials?.password}
            >
              <Field
                as={Input}
                name="credentials.password"
                autoComplete="off"
                type="password"
                id="ssh_password"
                maxLength={200}
                value={values.credentials.password}
                placeholder={isEditing ? '*******' : ''}
                data-cy="cloudSettings-SshPassword"
              />
            </FormControl>
            <SSHCredentialsPrivateKeyForm
              hasSSHKey={hasSSHKey}
              hasPassphrase={hasPassphrase}
              isEditing={isEditing}
              sshPrivateKeyValue={values.credentials.privateKey}
              privateKeyErrors={errors.credentials?.privateKey}
              passphraseValue={values.credentials.passphrase}
              passphraseErrors={errors.credentials?.passphrase}
            />
          </>

          <div className="form-group">
            <div className="col-sm-12 mt-3">
              <LoadingButton
                disabled={!isValid || !dirty}
                data-cy="createCredentials-saveButton"
                isLoading={isLoading}
                className="!ml-0"
                loadingText="Saving credentials..."
              >
                {isEditing ? 'Update credentials' : 'Add credentials'}
              </LoadingButton>
              {isEditing && (
                <Button
                  color="default"
                  onClick={() =>
                    router.stateService.go(
                      'portainer.settings.sharedcredentials'
                    )
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
