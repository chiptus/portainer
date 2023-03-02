import { useRouter } from '@uirouter/react';

import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { TextTip } from '@@/Tip/TextTip';

import {
  useCreateCredentialMutation,
  useCloudCredentials,
} from '../cloudSettings.service';
import {
  CreateCredentialPayload,
  GenericFormValues,
  credentialTitles,
  credentialTypeHelpLinks,
  CredentialType,
} from '../types';
import { APICredentialsForm } from '../components/APICredentialsForm';
import { GCPCredentialsForm } from '../components/GCPCredentialsForm';
import { AWSCredentialsForm } from '../components/AWSCredentialsForm';
import { AzureCredentialsForm } from '../components/AzureCredentialsForm';
import { SSHCredentialsForm } from '../components/SSHCredentialsForm';
import { trimObject } from '../utils';

type Props = {
  credentialType: CredentialType;
  routeOnSuccess?: string;
};

export function CredentialsForm({ credentialType, routeOnSuccess }: Props) {
  const router = useRouter();

  const Form = getForm(credentialType);
  const title = credentialTitles[credentialType];
  const helpLink = credentialTypeHelpLinks[credentialType];

  const createCredentialMutation = useCreateCredentialMutation();
  const cloudCredentialsQuery = useCloudCredentials();
  const credentialNames =
    cloudCredentialsQuery.data
      ?.filter((c) => c.provider === credentialType)
      .map((c: CreateCredentialPayload) => c.name) || [];

  return (
    <>
      <FormSectionTitle>Credential details</FormSectionTitle>
      <TextTip color="blue">
        <span>
          See our{' '}
          <a
            className="hyperlink"
            href={helpLink}
            target="_blank"
            rel="noreferrer"
          >
            documentation for obtaining {title} credentials.
          </a>{' '}
          Any credentials that you set up will be usable by all admin users{' '}
          (although the actual values themselves cannot be viewed).
        </span>
      </TextTip>
      <Form
        selectedProvider={credentialType}
        isLoading={createCredentialMutation.isLoading}
        onSubmit={onSubmit}
        credentialNames={credentialNames}
      />
    </>
  );

  function onSubmit(values: GenericFormValues) {
    const payload: CreateCredentialPayload = {
      provider: credentialType,
      name: values.name.trim(),
      credentials: trimObject(values.credentials),
    };
    // base64 encode the private key
    if (
      'privateKey' in payload.credentials &&
      typeof payload.credentials.privateKey === 'string'
    ) {
      payload.credentials.privateKey = window.btoa(
        payload.credentials.privateKey
      );
    }
    createCredentialMutation.mutate(payload, {
      onSuccess: () => {
        if (routeOnSuccess) {
          router.stateService.go(routeOnSuccess);
        }
      },
    });
  }
}

function getForm(credentialType: CredentialType) {
  switch (credentialType) {
    case CredentialType.GOOGLE_CLOUD:
      return GCPCredentialsForm;

    case CredentialType.AWS:
      return AWSCredentialsForm;

    case CredentialType.AZURE:
      return AzureCredentialsForm;

    case CredentialType.SSH:
      return SSHCredentialsForm;

    case CredentialType.CIVO:
    case CredentialType.DIGITAL_OCEAN:
    case CredentialType.LINODE:
    default:
      return APICredentialsForm;
  }
}
