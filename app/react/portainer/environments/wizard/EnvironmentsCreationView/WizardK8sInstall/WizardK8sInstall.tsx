import { useMemo, useState } from 'react';
import { Form, Formik } from 'formik';

import {
  credentialTypeToProvidersMap,
  CredentialType,
} from '@/react/portainer/settings/sharedCredentials/types';
import {
  useCloudCredentials,
  useCustomTemplates,
} from '@/react/portainer/settings/sharedCredentials/cloudSettings.service';
import { Environment } from '@/react/portainer/environments/types';
import { CredentialsForm } from '@/react/portainer/settings/sharedCredentials/CreateCredentialsView/CredentialsForm';
import Microk8s from '@/assets/ico/vendor/microk8s.svg?c';

import { Link } from '@@/Link';
import { TextTip } from '@@/Tip/TextTip';
import { BoxSelector } from '@@/BoxSelector';
import { Loading } from '@@/Widget';

import { AnalyticsStateKey } from '../types';

import { useInstallK8sCluster } from './service';
import { useValidationSchema } from './WizardK8sInstall.validation';
import { Microk8sCreateClusterForm } from './Microk8sCreateClusterForm';
import {
  K8sDistributionType,
  K8sInstallFormValues,
  k8sInstallTitles,
} from './types';
import { formatMicrok8sPayload } from './utils';

interface Props {
  onCreate(environment: Environment, analytics: AnalyticsStateKey): void;
}

const initialValues: K8sInstallFormValues = {
  name: '',
  credentialId: 0,
  meta: {
    groupId: 1,
    tagIds: [],
  },
  microk8s: {
    nodeIPs: [''],
    addons: [],
    customTemplateId: 0,
    kubernetesVersion: 'latest/stable',
  },
};

const k8sInstallOptions = [
  {
    id: K8sDistributionType.MICROK8S,
    icon: Microk8s,
    label: 'MicroK8s',
    description: 'Lightweight Kubernetes',
    value: K8sDistributionType.MICROK8S,
  },
];

export function WizardK8sInstall({ onCreate }: Props) {
  const [isSSHTestSuccessful, setIsSSHTestSuccessful] = useState<
    boolean | undefined
  >(undefined);
  const installK8sClusterMutation = useInstallK8sCluster();

  const [k8sDistributionType, setK8sDistributionType] =
    useState<K8sDistributionType>(K8sDistributionType.MICROK8S);

  const credentialType = CredentialType.SSH;

  const credentialsQuery = useCloudCredentials();
  const customTemplatesQuery = useCustomTemplates();

  const credentials = credentialsQuery.data;
  const availableCredentials = useMemo(
    () =>
      credentials?.filter((c) =>
        credentialTypeToProvidersMap[c.provider]?.includes(k8sDistributionType)
      ) || [],
    [credentials, k8sDistributionType]
  );

  const validation = useValidationSchema();
  const customTemplates =
    customTemplatesQuery.data?.filter((t) => t.Type === 3) || [];

  const isCredentialsFound = availableCredentials.length > 0;

  if (credentialsQuery.isLoading) {
    return <Loading />;
  }

  return (
    <>
      <Formik
        initialValues={initialValues}
        onSubmit={handleSubmit}
        validationSchema={validation}
        validateOnMount
        enableReinitialize
      >
        <Form className="form-horizontal">
          <BoxSelector
            radioName="k8sInstallForm-type"
            data-cy="k8sInstallForm-providerSelect"
            options={k8sInstallOptions}
            onChange={(provider: K8sDistributionType) =>
              setK8sDistributionType(provider)
            }
            value={k8sDistributionType}
          />
          {isCredentialsFound && (
            <Microk8sCreateClusterForm
              credentials={availableCredentials}
              customTemplates={customTemplates}
              isSubmitting={installK8sClusterMutation.isLoading}
              isSSHTestSuccessful={isSSHTestSuccessful}
              setIsSSHTestSuccessful={setIsSSHTestSuccessful}
            />
          )}
        </Form>
      </Formik>

      {!isCredentialsFound && (
        <>
          <TextTip color="orange">
            No SSH credentials found for {k8sInstallTitles[k8sDistributionType]}
            . Save your SSH credentials below, or in the{' '}
            <Link
              to="portainer.settings.sharedcredentials"
              title="shared credential settings"
              className="hyperlink"
            >
              shared credential settings.
            </Link>
          </TextTip>
          <CredentialsForm credentialType={credentialType} />
        </>
      )}
    </>
  );

  function handleSubmit(
    values: K8sInstallFormValues,
    {
      setFieldValue,
    }: {
      setFieldValue: (
        field: string,
        value: string | string[],
        shouldValidate?: boolean
      ) => void;
    }
  ) {
    const payload = formatMicrok8sPayload(values);

    installK8sClusterMutation.mutate(
      { payload, provider: k8sDistributionType },
      {
        onSuccess: (environment) => {
          onCreate(environment, 'k8sInstallAgent');
          setFieldValue('name', '');
          setFieldValue('microk8s.nodeIPs', ['']);
          setIsSSHTestSuccessful(undefined);
        },
      }
    );
  }
}
