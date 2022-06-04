import { Form, Formik } from 'formik';
import { useReducer, useState } from 'react';

import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { useCreateKubeConfigEnvironmentMutation } from '@/portainer/environments/queries/useCreateEnvironmentMutation';
import { notifySuccess } from '@/portainer/services/notifications';
import { Environment } from '@/portainer/environments/types';
import { CreateKubeConfigEnvironment } from '@/portainer/environments/environment.service/create';
import { FileUploadField } from '@/portainer/components/form-components/FileUpload/FileUploadField';
import { FormControl } from '@/portainer/components/form-components/FormControl';
import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';

import { NameField } from '../../shared/NameField';
import { MetadataFieldset } from '../../shared/MetadataFieldset';

import { validation } from './KubeConfig.validation';

interface Props {
  onCreate(environment: Environment): void;
}

const initialValues: CreateKubeConfigEnvironment = {
  kubeConfig: '',
  name: '',
  meta: {
    groupId: 1,
    tagIds: [],
  },
};

async function readFileContent(file: File) {
  return new Promise((resolve, reject) => {
    const fileReader = new FileReader();
    fileReader.onload = (e) => {
      if (e.target == null || e.target.result == null) {
        resolve('');
        return;
      }
      const base64 = e.target.result.toString();
      const index = base64.indexOf('base64,');
      // ignore first 7 characters (base64,)
      const cert = base64.substring(index + 7, base64.length);
      resolve(cert);
    };
    fileReader.onerror = () => {
      reject(new Error('error reading provisioning certificate file'));
    };
    fileReader.readAsDataURL(file);
  });
}

export function KubeConfigForm({ onCreate }: Props) {
  const [formKey, clearForm] = useReducer((state) => state + 1, 0);
  const [kubeConfigFile, setKubeConfigFile] = useState<File>();

  const mutation = useCreateKubeConfigEnvironmentMutation();

  async function handleFileUpload(
    file: File,
    setFieldValue: (
      field: string,
      value: unknown,
      shouldValidate?: boolean
    ) => void
  ) {
    if (file) {
      setKubeConfigFile(file);
      const fileContent = await readFileContent(file);
      setFieldValue('kubeConfig', fileContent);
    }
  }

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={handleSubmit}
      validationSchema={validation}
      validateOnMount
      key={formKey}
    >
      {({ isValid, dirty, setFieldValue, errors }) => (
        <Form>
          <FormSectionTitle>Environment details</FormSectionTitle>

          <div className="form-group">
            <div className="col-sm-12">
              <span className="text-primary">
                {' '}
                <i
                  className="fa fa-exclamation-circle"
                  aria-hidden="true"
                  style={{ marginRight: 2 }}
                />
              </span>
              <span className="text-muted small">
                Import the{' '}
                <a
                  href="https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/"
                  target="_blank"
                  rel="noreferrer"
                >
                  kubeconfig file
                </a>{' '}
                of an existing Kubernetes cluster located on-premise or on a
                cloud platform. This will create a corresponding environment in
                Portainer and install the agent on the cluster. Please ensure:
              </span>
            </div>
            <div className="col-sm-12 text-muted small">
              <ul style={{ padding: 10, paddingLeft: 20 }}>
                <li>You have a load balancer enabled in your cluster</li>
                <li>You specify current-context in your kubeconfig</li>
                <li>
                  The kubeconfig is self-contained - including any required
                  credentials.
                </li>
              </ul>
              <p>
                Note: Officially supported cloud providers are Civo, Linode,
                DigitalOcean and Microsoft Azure (others are not guaranteed to
                work at present)
              </p>
            </div>
          </div>

          <NameField />

          <FormControl
            label="Kubeconfig file"
            required
            errors={errors.kubeConfig}
            inputId="kubeconfig_file"
          >
            <FileUploadField
              inputId="kubeconfig_file"
              title="Select a file"
              accept=".yaml,.yml"
              value={kubeConfigFile}
              onChange={(file) => handleFileUpload(file, setFieldValue)}
            />
          </FormControl>

          <MetadataFieldset />

          <div className="form-group">
            <div className="col-sm-12">
              <LoadingButton
                className="wizard-connect-button"
                loadingText="Connecting environment..."
                isLoading={mutation.isLoading}
                disabled={!dirty || !isValid}
              >
                <i className="fa fa-plug" aria-hidden="true" /> Connect
              </LoadingButton>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );

  function handleSubmit(values: CreateKubeConfigEnvironment) {
    mutation.mutate(values, {
      onSuccess(environment) {
        notifySuccess('Kubeconfig import started', environment.Name);
        clearForm();
        setKubeConfigFile(undefined);
        onCreate(environment);
      },
    });
  }
}
