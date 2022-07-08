import { useRouter } from '@uirouter/react';

import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { TextTip } from '@@/Tip/TextTip';

import {
  useCreateCredentialMutation,
  useCloudCredentials,
} from '../cloudSettings.service';
import {
  CreateCredentialPayload,
  KaasProvider,
  providerTitles,
  providerHelpLinks,
  GenericFormValues,
} from '../types';
import { APICredentialsForm } from '../components/APICredentialsForm';
import { GCPCredentialsForm } from '../components/GCPCredentialsForm';
import { AWSCredentialsForm } from '../components/AWSCredentialsForm';
import { AzureCredentialsForm } from '../components/AzureCredentialsForm';
import { trimObject } from '../utils';

type Props = {
  selectedProvider: KaasProvider;
  routeOnSuccess?: string;
};

export function CredentialsForm({ selectedProvider, routeOnSuccess }: Props) {
  const router = useRouter();

  const Form = getForm(selectedProvider);
  const title = providerTitles[selectedProvider];
  const helpLink = providerHelpLinks[selectedProvider];

  const createCredentialMutation = useCreateCredentialMutation();
  const cloudCredentialsQuery = useCloudCredentials();
  const credentialNames =
    cloudCredentialsQuery.data
      ?.filter((c) => c.provider === selectedProvider)
      .map((c: CreateCredentialPayload) => c.name) || [];

  return (
    <>
      <FormSectionTitle>Credential details</FormSectionTitle>
      <TextTip color="blue">
        See our{' '}
        <a href={helpLink} target="_blank" rel="noreferrer">
          documentation for obtaining {title} credentials.
        </a>{' '}
        Any credentials that you set up will be usable by all admin users{' '}
        (although the actual values themselves cannot be viewed).
      </TextTip>
      <Form
        selectedProvider={selectedProvider}
        isLoading={createCredentialMutation.isLoading}
        onSubmit={onSubmit}
        credentialNames={credentialNames}
      />
    </>
  );

  function onSubmit(values: GenericFormValues) {
    const payload: CreateCredentialPayload = {
      provider: selectedProvider,
      name: values.name.trim(),
      credentials: trimObject(values.credentials),
    };
    createCredentialMutation.mutate(payload, {
      onSuccess: () => {
        if (routeOnSuccess) {
          router.stateService.go(routeOnSuccess);
        }
      },
    });
  }
}

function getForm(provider: KaasProvider) {
  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      return GCPCredentialsForm;

    case KaasProvider.AWS:
      return AWSCredentialsForm;

    case KaasProvider.AZURE:
      return AzureCredentialsForm;

    case KaasProvider.CIVO:
    case KaasProvider.DIGITAL_OCEAN:
    case KaasProvider.LINODE:
    default:
      return APICredentialsForm;
  }
}
