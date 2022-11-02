import { useMemo, useState } from 'react';
import { Form, Formik } from 'formik';

import { react2angular } from '@/react-tools/react2angular';
import {
  KaasProvider,
  Credential,
  providerTitles,
} from '@/react/portainer/settings/cloud/types';
import { useCloudCredentials } from '@/react/portainer/settings/cloud/cloudSettings.service';
import { Environment } from '@/react/portainer/environments/types';
import { useSettings } from '@/react/portainer/settings/queries';
import { CredentialsForm } from '@/react/portainer/settings/cloud/CreateCredentialsView/CredentialsForm';

import { Loading } from '@@/Widget/Loading';
import { Link } from '@@/Link';
import { TextTip } from '@@/Tip/TextTip';

import { AnalyticsStateKey } from '../types';

import { KaasProvidersSelector } from './KaasProvidersSelector';
import { sendKaasProvisionAnalytics } from './utils';
import { useCloudProviderOptions, useCreateKaasCluster } from './queries';
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
  const createKaasClusterMutation = useCreateKaasCluster();

  const [provider, setProvider] = useState<KaasProvider>(KaasProvider.CIVO);
  const [credential, setCredential] = useState<Credential | null>(null);

  const credentialsQuery = useCloudCredentials();

  const cloudOptionsQuery = useCloudProviderOptions<KaasInfo>(
    provider,
    isKaasInfo,
    credential
  );

  const credentials = credentialsQuery.data;

  const providerCredentials = useMemo(
    () => credentials?.filter((c) => c.provider === provider) || [],
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
              onChangeSelectedCredential={setCredential}
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
            {providerTitles[provider]} credentials below, or in the
            <Link
              to="portainer.settings.cloud"
              title="cloud settings"
              className="hyperlink"
            >
              cloud settings.
            </Link>
          </TextTip>
          <CredentialsForm selectedProvider={provider} />
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
        value: string,
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
        },
      }
    );
  }
}

export const KaasFormGroupAngular = react2angular(WizardKaaS, ['onCreate']);
function isKaasInfo(value: KaasInfo): value is KaasInfo {
  return true;
}
