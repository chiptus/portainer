import { useMemo, useState } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';
import { Form, Formik } from 'formik';

import { react2angular } from '@/react-tools/react2angular';
import {
  KaasProvider,
  Credential,
  providerTitles,
} from '@/portainer/settings/cloud/types';
import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';
import { useCloudCredentials } from '@/portainer/settings/cloud/cloudSettings.service';
import { Environment } from '@/portainer/environments/types';
import { AnalyticsStateKey } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/types';
import { NameField } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';
import { useSettings } from '@/portainer/settings/queries';
import { FormSection } from '@/portainer/components/form-components/FormSection';
import { TextTip } from '@/portainer/components/Tip/TextTip';
import { Link } from '@/portainer/components/Link';
import { CredentialsForm } from '@/portainer/settings/cloud/CreateCredentialsView/CredentialsForm';
import { Loading } from '@/portainer/components/widget/Loading';

import { KaasProvidersSelector } from './KaasProvidersSelector';
import { sendKaasProvisionAnalytics } from './utils';
import { useCloudProviderOptions, useCreateKaasCluster } from './queries';
import { validationSchema } from './KaaSEnvironmentForm.validation';
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
  nodeSize: '',
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
  },
  api: {
    networkId: '',
  },
  azure: {
    resourceGroup: '',
    resourceGroupName: '',
    tier: 'Free',
    poolName: '',
    dnsPrefix: '',
    availabilityZones: [],
    resourceGroupInput: 'select',
  },
  amazon: {
    amiType: '',
    instanceType: '',
    nodeVolumeSize: 20,
  },
};

export function KaaSFormGroup({ onCreate }: Props) {
  const settingsQuery = useSettings();
  const createKaasClusterMutation = useCreateKaasCluster();

  const [provider, setProvider] = useState<KaasProvider>(KaasProvider.CIVO);
  const [credential, setCredential] = useState<Credential | null>(null);
  const { state } = useCurrentStateAndParams();

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

  const credentialsFound = providerCredentials.length > 0;

  return (
    <>
      <Formik
        initialValues={initialValues}
        onSubmit={handleSubmit}
        validationSchema={() =>
          validationSchema(provider, cloudOptionsQuery.data)
        }
        validateOnMount
        enableReinitialize
      >
        <Form className="form-horizontal">
          {state.name === 'portainer.endpoints.new' && (
            <FormSectionTitle>Environment details</FormSectionTitle>
          )}

          <NameField
            tooltip="Name of the cluster and environment"
            placeholder="e.g. my-cluster-name"
          />

          <FormSection title="Cluster details">
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
          </FormSection>
        </Form>
      </Formik>

      {!credentialsFound && (
        <>
          <TextTip color="orange">
            No API key found for
            <span className="mx-1">{providerTitles[provider]}.</span>
            Save your
            <span className="mx-1">{providerTitles[provider]}</span>
            credentials below, or in the
            <Link
              to="portainer.settings.cloud"
              title="cloud settings"
              className="ml-1"
            >
              cloud settings
            </Link>
            .
          </TextTip>
          <CredentialsForm selectedProvider={provider} />
        </>
      )}
    </>
  );

  function handleSubmit(
    values: FormValues,
    { resetForm }: { resetForm: () => void }
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
          resetForm();
        },
      }
    );
  }
}

export const KaasFormGroupAngular = react2angular(KaaSFormGroup, ['onCreate']);

function isKaasInfo(value: KaasInfo): value is KaasInfo {
  return true;
}
