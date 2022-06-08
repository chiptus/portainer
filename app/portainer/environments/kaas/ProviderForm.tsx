import { useFormikContext } from 'formik';
import { useEffect } from 'react';

import { KaasProvider, Credential } from '@/portainer/settings/cloud/types';

import { FormValues } from './types';
import { GKECreateClusterForm } from './GKECreateClusterForm/GKECreateClusterForm';
import { ApiCreateClusterForm } from './ApiCreateClusterForm/ApiCreateClusterForm';
import { AzureCreateClusterForm } from './AzureCreateClusterForm/AzureCreateClusterForm';
import { EKSCreateClusterForm } from './EKSCreateClusterForm/EKSCreateClusterForm';

interface Props {
  provider: KaasProvider;
  onChangeSelectedCredential: (credential: Credential | null) => void;
  credentials: Credential[];
  isSubmitting: boolean;
}

export function ProviderForm({
  provider,
  onChangeSelectedCredential,
  credentials,

  isSubmitting,
}: Props) {
  useSelectedCredentials(credentials, onChangeSelectedCredential);

  const Form = getForm(provider);

  if (credentials.length === 0) {
    return null;
  }

  return (
    <Form
      provider={provider}
      credentials={credentials}
      isSubmitting={isSubmitting}
    />
  );
}

// to expand when other create cluster forms are added
function getForm(provider: KaasProvider) {
  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      return GKECreateClusterForm;
    case KaasProvider.CIVO:
    case KaasProvider.DIGITAL_OCEAN:
    case KaasProvider.LINODE:
      return ApiCreateClusterForm;
    case KaasProvider.AZURE:
      return AzureCreateClusterForm;
    case KaasProvider.AWS:
      return EKSCreateClusterForm;
    default:
      throw new Error(`Provider ${provider} not supported`);
  }
}

function useSelectedCredentials(
  credentials: Credential[],
  onChange: (credential: Credential | null) => void
) {
  const { values } = useFormikContext<FormValues>();

  const selectedCredential = credentials.length
    ? credentials.find((c) => c.id === values.credentialId) || credentials[0]
    : null;

  useEffect(() => {
    onChange(selectedCredential);
  }, [onChange, selectedCredential]);
}
