import { Field, Form, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { Button } from '@@/buttons';

import { GitCredentialFormValues } from '../types';

import { validationSchema } from './GitCredentialForm.validation';

type Props = {
  isEditing?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: GitCredentialFormValues) => void;
  gitCredentialNames: string[];
  initialValues?: GitCredentialFormValues;
};

const defaultInitialValues = {
  name: '',
  username: '',
  password: '',
};

export function GitCredentialForm({
  isEditing = false,
  isLoading,
  onSubmit,
  gitCredentialNames,
  initialValues = defaultInitialValues,
}: Props) {
  const router = useRouter();

  return (
    <Formik
      initialValues={initialValues}
      enableReinitialize
      validationSchema={() => validationSchema(gitCredentialNames, isEditing)}
      onSubmit={(values) => onSubmit(values)}
      validateOnMount
    >
      {({ values, errors, handleSubmit, isValid, dirty }) => (
        <Form className="form-horizontal" onSubmit={handleSubmit} noValidate>
          <FormControl
            inputId="Name"
            label="Name"
            errors={errors.name}
            required
          >
            <Field
              as={Input}
              name="name"
              value={values.name}
              autoComplete="off"
              id="Name"
            />
          </FormControl>

          <FormControl
            inputId="Username"
            label="Username"
            errors={errors.username}
          >
            <Field
              as={Input}
              name="username"
              value={values.username}
              autoComplete="off"
              id="Username"
            />
          </FormControl>

          <FormControl
            inputId="Password"
            label="Personal Access Token"
            errors={errors.password}
            required={!isEditing}
          >
            <Field
              as={Input}
              name="password"
              value={values.password}
              autoComplete="off"
              id="Password"
            />
          </FormControl>
          <div className="form-group">
            <div className="col-sm-12 mt-3">
              <LoadingButton
                disabled={!isValid || !dirty}
                isLoading={isLoading}
                loadingText="Saving Git Credential..."
              >
                {isEditing ? 'Update git credential' : 'Save git credential'}
              </LoadingButton>
              {isEditing && (
                <Button
                  color="default"
                  onClick={() => router.stateService.go('portainer.account')}
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
