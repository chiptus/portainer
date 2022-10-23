import { Form, Formik } from 'formik';
import { useReducer, useState } from 'react';

import { useCreateKubeConfigEnvironmentMutation } from '@/react/portainer/environments/queries/useCreateEnvironmentMutation';
import { notifySuccess } from '@/portainer/services/notifications';
import { Environment } from '@/react/portainer/environments/types';
import { CreateKubeConfigEnvironment } from '@/react/portainer/environments/environment.service/create';

import { FormControl } from '@@/form-components/FormControl';
import { FileUploadField } from '@@/form-components/FileUpload/FileUploadField';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { Icon } from '@@/Icon';
import { TextTip } from '@@/Tip/TextTip';

import { NameField } from '../../shared/NameField';
import { MoreSettingsSection } from '../../shared/MoreSettingsSection';

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
          <div className="form-group">
            <div className="col-sm-12">
              <TextTip color="blue">
                <span className="text-muted">
                  <a
                    href="https://docs.portainer.io/start/install/agent/kubernetes/import"
                    target="_blank"
                    rel="noreferrer"
                    className="mx-1"
                  >
                    Import the kubeconfig file
                  </a>
                  of an existing Kubernetes cluster located on-premise or on a
                  cloud platform. This will create a corresponding environment
                  in Portainer and install the agent on the cluster. Please
                  ensure:
                </span>
              </TextTip>
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

          <MoreSettingsSection />

          <div className="form-group">
            <div className="col-sm-12">
              <LoadingButton
                className="wizard-connect-button vertical-center"
                loadingText="Connecting environment..."
                isLoading={mutation.isLoading}
                disabled={!dirty || !isValid}
              >
                <Icon
                  icon="svg-plug"
                  className="icon icon-sm vertical-center"
                />{' '}
                Connect
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
