import { useFormikContext } from 'formik';
import { useEffect } from 'react';

import {
  KaasProvider,
  Credential,
  CustomTemplate,
} from '@/react/portainer/settings/cloud/types';

import { FormValues } from './types';
import { GKECreateClusterForm } from './GKECreateClusterForm/GKECreateClusterForm';
import { ApiCreateClusterForm } from './ApiCreateClusterForm/ApiCreateClusterForm';
import { AzureCreateClusterForm } from './AzureCreateClusterForm/AzureCreateClusterForm';
import { EKSCreateClusterForm } from './EKSCreateClusterForm/EKSCreateClusterForm';
import { Microk8sCreateClusterForm } from './Microk8sCreateClusterForm/Microk8sCreateClusterForm';

interface Props {
  provider: KaasProvider;
  onChangeSelectedCredential: (credential: Credential | null) => void;
  credentials: Credential[];
  customTemplates: CustomTemplate[];
  isSubmitting: boolean;
}

export function ProviderForm({
  provider,
  onChangeSelectedCredential,
  credentials,
  customTemplates,
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
      customTemplates={customTemplates}
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
    case KaasProvider.MICROK8S:
      return Microk8sCreateClusterForm;
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
