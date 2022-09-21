import _ from 'lodash';
import { useRouter } from '@uirouter/react';

import {
  AccessKeyFormValues,
  APIFormValues,
  AzureFormValues,
  Credential,
  CreateCredentialPayload,
  GenericFormValues,
  KaasProvider,
  ServiceAccountFormValues,
  UpdateCredentialPayload,
} from '../types';
import { APICredentialsForm } from '../components/APICredentialsForm';
import { GCPCredentialsForm } from '../components/GCPCredentialsForm';
import { AWSCredentialsForm } from '../components/AWSCredentialsForm';
import { AzureCredentialsForm } from '../components/AzureCredentialsForm';
import { sensitiveFieldChanged, sensitiveFields, trimObject } from '../utils';
import {
  useCloudCredentials,
  useUpdateCredentialMutation,
} from '../cloudSettings.service';

type Props = {
  credential: Credential;
};

enum FormTypes {
  API = 'api',
  SERVICE_ACCOUNT = 'service-account',
  ACCESS_KEY = 'access-key',
  AZURE = 'azure',
}

export function EditCredentialForm({ credential }: Props) {
  const { provider } = credential;
  const formType = getFormType(provider);

  const router = useRouter();
  const cloudCredentialsQuery = useCloudCredentials();
  const credentialNames =
    cloudCredentialsQuery.data
      ?.filter((c) => c.name !== credential.name && c.provider === provider)
      .map((c: CreateCredentialPayload) => c.name) || [];

  const updateCredentialMutation = useUpdateCredentialMutation();

  const initialValues = {
    name: credential.name,
    provider: credential.provider,
    credentials: credential.credentials,
  };

  return (
    <>
      {formType === FormTypes.API && 'apiKey' in credential.credentials && (
        <APICredentialsForm
          selectedProvider={provider}
          isEditing
          isLoading={updateCredentialMutation.isLoading}
          onSubmit={onSubmit}
          credentialNames={credentialNames}
          initialValues={
            {
              ...initialValues,
              credentials: { ...initialValues.credentials, apiKey: '' },
            } as APIFormValues
          }
          placeholderText="*******"
        />
      )}
      {formType === FormTypes.SERVICE_ACCOUNT &&
        'jsonKeyBase64' in credential.credentials && (
          <GCPCredentialsForm
            selectedProvider={provider}
            isEditing
            isLoading={updateCredentialMutation.isLoading}
            onSubmit={onSubmit}
            credentialNames={credentialNames}
            initialValues={
              {
                ...initialValues,
                credentials: {
                  ...initialValues.credentials,
                  jsonKeyBase64: '',
                },
              } as ServiceAccountFormValues
            }
          />
        )}
      {formType === FormTypes.ACCESS_KEY &&
        'secretAccessKey' in credential.credentials && (
          <AWSCredentialsForm
            selectedProvider={provider}
            isEditing
            isLoading={updateCredentialMutation.isLoading}
            onSubmit={onSubmit}
            credentialNames={credentialNames}
            initialValues={
              {
                ...initialValues,
                credentials: {
                  ...initialValues.credentials,
                  secretAccessKey: '',
                },
              } as AccessKeyFormValues
            }
            placeholderText="*******"
          />
        )}
      {formType === FormTypes.AZURE &&
        'clientSecret' in credential.credentials && (
          <AzureCredentialsForm
            selectedProvider={provider}
            isEditing
            isLoading={updateCredentialMutation.isLoading}
            onSubmit={onSubmit}
            credentialNames={credentialNames}
            initialValues={
              {
                ...initialValues,
                credentials: { ...initialValues.credentials, clientSecret: '' },
              } as AzureFormValues
            }
            placeholderText="*******"
          />
        )}
    </>
  );

  function onSubmit(values: GenericFormValues) {
    const newCredentials: UpdateCredentialPayload = {
      name: values.name.trim(),
      provider,
    };
    if (sensitiveFieldChanged(values.credentials)) {
      newCredentials.credentials = trimObject(values.credentials);
    } else {
      const visibleCredentials = _.omit(values.credentials, sensitiveFields);
      if (Object.keys(visibleCredentials).length > 0) {
        newCredentials.credentials = trimObject(visibleCredentials);
      }
    }
    updateCredentialMutation.mutate({
      credential: newCredentials,
      id: credential.id,
    });
    router.stateService.go('portainer.settings.cloud');
  }
}

function getFormType(provider: KaasProvider) {
  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      return FormTypes.SERVICE_ACCOUNT;

    case KaasProvider.AWS:
      return FormTypes.ACCESS_KEY;

    case KaasProvider.AZURE:
      return FormTypes.AZURE;

    case KaasProvider.CIVO:
    case KaasProvider.DIGITAL_OCEAN:
    case KaasProvider.LINODE:
    default:
      return FormTypes.API;
  }
}
