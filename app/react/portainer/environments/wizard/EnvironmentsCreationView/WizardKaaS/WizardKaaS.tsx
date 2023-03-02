import { useMemo, useState } from 'react';
import { Form, Formik } from 'formik';

import {
  Credential,
  credentialTypeToProvidersMap,
  providerToCredentialTypeMap,
} from '@/react/portainer/settings/sharedCredentials/types';
import { useCloudCredentials } from '@/react/portainer/settings/sharedCredentials/cloudSettings.service';
import { Environment } from '@/react/portainer/environments/types';
import { useSettings } from '@/react/portainer/settings/queries';
import { CredentialsForm } from '@/react/portainer/settings/sharedCredentials/CreateCredentialsView/CredentialsForm';

import { Loading } from '@@/Widget/Loading';
import { Link } from '@@/Link';
import { TextTip } from '@@/Tip/TextTip';

import { AnalyticsStateKey } from '../types';
import { KaasProvider, providerTitles } from '../WizardK8sInstall/types';

import { KaasProvidersSelector } from './KaasProvidersSelector';
import { sendKaasProvisionAnalytics } from './utils';
import { useCloudProviderOptions, useCreateCluster } from './queries';
import { useValidationSchema } from './WizardKaaS.validation';
import { ProviderForm } from './ProviderForm';
import { FormValues, KaasInfo } from './types';
import { getPayloadParse } from './converter';

interface Props {
  onCreate(environment: Environment, analytics: AnalyticsStateKey): void;
}

const initialValues: FormValues = {
  name: '',
  nodeCount: 3,
  kubernetesVersion: '',
  region: '',
  credentialId: 0,

  meta: {
    groupId: 1,
    tagIds: [],
  },

  google: {
    cpu: 2,
    ram: 4,
    hdd: 100,
    networkId: '',
    nodeSize: '',
  },
  api: {
    networkId: '',
    nodeSize: '',
  },
  azure: {
    resourceGroup: '',
    resourceGroupName: '',
    tier: 'Free',
    poolName: '',
    dnsPrefix: '',
    availabilityZones: [],
    resourceGroupInput: 'select',
    nodeSize: '',
  },
  amazon: {
    amiType: '',
    instanceType: '',
    nodeVolumeSize: 20,
  },
};

export function WizardKaaS({ onCreate }: Props) {
  const settingsQuery = useSettings();
  const createKaasClusterMutation = useCreateCluster();

  const [provider, setProvider] = useState<KaasProvider>(KaasProvider.CIVO);
  const [credentialType, setCredentialType] = useState<Credential | null>(null);

  const credentialsQuery = useCloudCredentials();

  const cloudOptionsQuery = useCloudProviderOptions<KaasInfo>(
    provider,
    isKaasInfo,
    credentialType
  );

  const credentials = credentialsQuery.data;

  const providerCredentials = useMemo(
    () =>
      credentials?.filter((c) =>
        // use only the credentials that have a type that support the selected provider
        credentialTypeToProvidersMap[c.provider]?.includes(provider)
      ) || [],
    [credentials, provider]
  );

  const validation = useValidationSchema(provider, cloudOptionsQuery.data);

  const credentialsFound = providerCredentials.length > 0;

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
          <KaasProvidersSelector provider={provider} onChange={setProvider} />

          {credentialsQuery.isLoading ? (
            <Loading />
          ) : (
            <ProviderForm
              provider={provider}
              onChangeSelectedCredential={setCredentialType}
              credentials={providerCredentials}
              isSubmitting={createKaasClusterMutation.isLoading}
            />
          )}
        </Form>
      </Formik>

      {!credentialsFound && (
        <>
          <TextTip color="orange">
            No API key found for {providerTitles[provider]}. Save your{' '}
            {providerTitles[provider]} credentials below, or in the{' '}
            <Link
              to="portainer.settings.sharedcredentials"
              title="shared credential settings"
              className="hyperlink"
            >
              shared credential settings.
            </Link>
          </TextTip>
          <CredentialsForm
            credentialType={providerToCredentialTypeMap[provider]}
          />
        </>
      )}
    </>
  );

  function handleSubmit(
    values: FormValues,
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
    if (settingsQuery.data?.EnableTelemetry) {
      sendKaasProvisionAnalytics(values, provider);
    }

    const parser = getPayloadParse(provider);

    const payload = parser(values);

    createKaasClusterMutation.mutate(
      { payload, provider },
      {
        onSuccess: (environment) => {
          onCreate(environment, 'kaasAgent');
          setFieldValue('name', '');
          setFieldValue('microk8s.nodeIPs', ['']);
        },
      }
    );
  }
}

function isKaasInfo(value: KaasInfo): value is KaasInfo {
  return true;
}
