import { useState } from 'react';
import { Field, Form, Formik } from 'formik';

import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input } from '@/portainer/components/form-components/Input';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { FileUploadField } from '@/portainer/components/form-components/FileUpload';
import { error as notifyError } from '@/portainer/services/notifications';
import { readFileAsArrayBuffer } from '@/portainer/services/fileUploadReact';

import {
  ServiceAccountFormValues,
  KaasProvider,
  providerTitles,
} from '../types';
import { isMeaningfulChange } from '../utils';

import { validationSchema } from './GCPCredentialsForm.validation';

type Props = {
  selectedProvider: KaasProvider;
  showProviderInput?: boolean;
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
  showProviderInput = false,
  isLoading,
  onSubmit,
  credentialNames,
  initialValues = defaultInitialValues,
}: Props) {
  const [serviceKeyFile, setserviceKeyFile] = useState<File>();

  return (
    <Formik
      initialValues={initialValues}
      enableReinitialize
      validationSchema={() => validationSchema(credentialNames)}
      onSubmit={(values) => onSubmit(values)}
      validateOnMount
    >
      {({ values, errors, handleSubmit, setFieldValue, isValid, dirty }) => (
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
          <FormControl inputId="name" label="Name" errors={errors.name}>
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
            label="Service Account Key (.json)"
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
