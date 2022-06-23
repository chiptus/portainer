import { useState } from 'react';
import { Field, Form, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { error as notifyError } from '@/portainer/services/notifications';
import { readFileAsArrayBuffer } from '@/portainer/services/fileUploadReact';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { FileUploadField } from '@@/form-components/FileUpload';
import { Button } from '@@/buttons';

import {
  ServiceAccountFormValues,
  KaasProvider,
  providerTitles,
} from '../types';

import { validationSchema } from './GCPCredentialsForm.validation';

type Props = {
  selectedProvider: KaasProvider;
  isEditing?: boolean;
  isLoading: boolean;
  onSubmit: (formValues: ServiceAccountFormValues) => void;
  credentialNames: string[];
  initialValues?: ServiceAccountFormValues;
};

const defaultInitialValues = {
  name: '',
  credentials: {
    jsonKeyBase64: '',
  },
};

export function GCPCredentialsForm({
  selectedProvider,
  isEditing = false,
  isLoading,
  onSubmit,
  credentialNames,
  initialValues = defaultInitialValues,
}: Props) {
  const [serviceKeyFile, setserviceKeyFile] = useState<File>();
  const router = useRouter();

  return (
    <Formik
      initialValues={initialValues}
      enableReinitialize
      validationSchema={() => validationSchema(credentialNames, isEditing)}
      onSubmit={(values) => onSubmit(values)}
      validateOnMount
    >
      {({ values, errors, handleSubmit, setFieldValue, isValid, dirty }) => (
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
            inputId="credentials.jsonKeyBase64"
            label="Service account key (.json)"
            tooltip="Service account keys can be created in the 'IAM and admin' section in Google Cloud Console"
            errors={errors.credentials?.jsonKeyBase64}
          >
            <FileUploadField
              inputId="credentials.jsonKeyBase64"
              title="Upload file"
              accept=".json"
              value={serviceKeyFile}
              onChange={(file) => handleFileUpload(file, setFieldValue)}
            />
          </FormControl>

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

  async function handleFileUpload(
    file: File,
    setFieldValue: (
      field: string,
      value: unknown,
      shouldValidate?: boolean
    ) => void
  ) {
    if (file) {
      setserviceKeyFile(file);
      try {
        const arrayBufferContent = await readFileAsArrayBuffer(file);
        if (arrayBufferContent && typeof arrayBufferContent !== 'string') {
          const base64Content = arrayBufferToBase64(arrayBufferContent);
          setFieldValue('credentials.jsonKeyBase64', base64Content);
        } else {
          notifyError('Unable to read the uploaded file');
        }
      } catch (error) {
        notifyError('Unable to read the uploaded file');
      }
    }
  }
}

function arrayBufferToBase64(buffer: ArrayBuffer) {
  let binary = '';
  const bytes = new Uint8Array(buffer);
  const len = bytes.byteLength;
  for (let i = 0; i < len; i += 1) {
    binary += String.fromCharCode(bytes[i]);
  }
  return window.btoa(binary);
}
