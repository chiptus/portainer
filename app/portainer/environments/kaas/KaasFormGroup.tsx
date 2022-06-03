import { useEffect, useState } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { react2angular } from '@/react-tools/react2angular';
import { KaasProvider, providerTitles } from '@/portainer/settings/cloud/types';
import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';
import { useCloudCredentials } from '@/portainer/settings/cloud/cloudSettings.service';
import { CredentialsForm } from '@/portainer/settings/cloud/CreateCredentialsView/CredentialsForm';
import { Loading } from '@/portainer/components/widget/Loading';
import { Alert } from '@/portainer/components/Alert/Alert';
import { Link } from '@/portainer/components/Link';
import { Environment } from '@/portainer/environments/types';
import { AnalyticsStateKey } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/types';

import { KaasProvidersSelector } from './KaasProvidersSelector';
import { EnvironmentNameForm } from './EnvironmentNameForm/EnvironmentNameForm';
import { ApiCreateClusterFormContainer } from './ApiCreateClusterForm/ApiCreateClusterFormContainer';
import { GKECreateClusterFormContainer } from './GKECreateClusterForm/GKECreateClusterFormContainer';
import { AzureCreateClusterFormContainer } from './AzureCreateClusterForm/AzureCreateClusterFormContainer';
import { EKSCreateClusterFormContainer } from './EKSCreateClusterForm/EKSCreateClusterFormContainer';

interface Props {
  onCreate?(environment: Environment, analytics: AnalyticsStateKey): void;
}

export function KaasFormGroup({ onCreate }: Props) {
  const [selectedProvider, setSelectedProvider] = useState<KaasProvider>(
    KaasProvider.CIVO
  );
  const [environmentName, setEnvironmentName] = useState('');

  const { state } = useCurrentStateAndParams();

  // select an initial provider that has credentials available
  const credentialsQuery = useCloudCredentials();
  useEffect(() => {
    if (credentialsQuery.data && credentialsQuery.data.length > 0) {
      const credentialAvailable = credentialsQuery.data.some(
        (credential) => credential.provider === selectedProvider
      );
      if (!credentialAvailable) {
        setSelectedProvider(credentialsQuery.data[0].provider);
      }
      // do nothing if there is a credential available
      return;
    }
    setSelectedProvider(KaasProvider.CIVO);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [credentialsQuery.data]);

  const providerCredentials =
    credentialsQuery.data?.filter((c) => c.provider === selectedProvider) || [];

  const CreateClusterForm = getForm(selectedProvider);

  return (
    <>
      {state.name === 'portainer.endpoints.new' && (
        <FormSectionTitle>Environment details</FormSectionTitle>
      )}
      <EnvironmentNameForm
        environmentName={environmentName}
        setEnvironmentName={setEnvironmentName}
      />
      <FormSectionTitle>Cluster details</FormSectionTitle>
      <KaasProvidersSelector
        provider={selectedProvider}
        onChange={(provider) => {
          setSelectedProvider(provider);
        }}
      />
      {/* // switch between the create cluster forms based on the selected provider */}
      {providerCredentials.length >= 1 && (
        <CreateClusterForm
          name={environmentName}
          setName={setEnvironmentName}
          provider={selectedProvider}
          credentials={providerCredentials}
          onCreate={onCreate}
        />
      )}
      {credentialsQuery.isLoading && <Loading />}
      {credentialsQuery.data && providerCredentials.length === 0 && (
        <>
          <Alert>
            No API key found for {providerTitles[selectedProvider]}. Save your{' '}
            {providerTitles[selectedProvider]} credentials below, or in
            the&nbsp;
            <Link to="portainer.settings.cloud" title="cloud settings">
              cloud settings
            </Link>
            .
          </Alert>
          <CredentialsForm selectedProvider={selectedProvider} />
        </>
      )}
    </>
  );
}

// to expand when other create cluster forms are added
function getForm(provider: KaasProvider) {
  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      return GKECreateClusterFormContainer;
    case KaasProvider.AZURE:
      return AzureCreateClusterFormContainer;
    case KaasProvider.AWS:
      return EKSCreateClusterFormContainer;
    case KaasProvider.CIVO:
    case KaasProvider.DIGITAL_OCEAN:
    case KaasProvider.LINODE:
    default:
      return ApiCreateClusterFormContainer;
  }
}

export const KaasFormGroupAngular = react2angular(KaasFormGroup, ['onCreate']);
