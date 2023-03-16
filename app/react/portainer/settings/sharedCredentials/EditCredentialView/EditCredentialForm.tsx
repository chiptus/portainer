import _ from 'lodash';
import { useRouter } from '@uirouter/react';

import {
  AccessKeyFormValues,
  APIFormValues,
  AzureFormValues,
  SSHCredentialFormValues,
  Credential,
  CreateCredentialPayload,
  GenericFormValues,
  ServiceAccountFormValues,
  UpdateCredentialPayload,
  CredentialType,
} from '../types';
import { APICredentialsForm } from '../components/APICredentialsForm';
import { GCPCredentialsForm } from '../components/GCPCredentialsForm';
import { AWSCredentialsForm } from '../components/AWSCredentialsForm';
import { AzureCredentialsForm } from '../components/AzureCredentialsForm';
import { SSHCredentialsForm } from '../components/SSHCredentialsForm/SSHCredentialsForm';
import { getUnchangedSensitiveFields, trimObject } from '../utils';
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
  USERNAME_PASSWORD = 'username-password',
}

export function EditCredentialForm({ credential }: Props) {
  const { provider: credentialType } = credential;
  const formType = getFormType(credentialType);

  const router = useRouter();
  const cloudCredentialsQuery = useCloudCredentials();
  const credentialNames =
    cloudCredentialsQuery.data
      ?.filter(
        (c) => c.name !== credential.name && c.provider === credentialType
      )
      .map((c: CreateCredentialPayload) => c.name) || [];

  const updateCredentialMutation = useUpdateCredentialMutation(credential.id);

  const initialValues = {
    name: credential.name,
    provider: credential.provider,
    credentials: credential.credentials,
  };

  return (
    <>
      {formType === FormTypes.API && 'apiKey' in credential.credentials && (
        <APICredentialsForm
          selectedProvider={credentialType}
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
            selectedProvider={credentialType}
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
            selectedProvider={credentialType}
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
            selectedProvider={credentialType}
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
      {formType === FormTypes.USERNAME_PASSWORD &&
        'password' in credential.credentials && (
          <SSHCredentialsForm
            selectedProvider={credentialType}
            isEditing
            isLoading={updateCredentialMutation.isLoading}
            onSubmit={onSubmit}
            credentialNames={credentialNames}
            hasSSHKey={!!credential.credentials.privateKey}
            hasPassphrase={!!credential.credentials.passphrase}
            initialValues={
              {
                ...initialValues,
                credentials: {
                  ...initialValues.credentials,
                  password: '',
                  privateKey: '',
                  passphrase: '',
                },
              } as SSHCredentialFormValues
            }
          />
        )}
    </>
  );

  function onSubmit(values: GenericFormValues) {
    const newCredentials: UpdateCredentialPayload = {
      name: values.name.trim(),
      provider: credentialType,
    };
    const unchangedSensitiveFields = getUnchangedSensitiveFields(
      values.credentials
    );
    // keep the unchanged sensitive fields out of put request payload
    const visibleCredentials = _.omit(
      values.credentials,
      unchangedSensitiveFields
    );
    if (Object.keys(visibleCredentials).length > 0) {
      newCredentials.credentials = trimObject(visibleCredentials);
    }
    // base64 encode the private key
    if (
      newCredentials.credentials &&
      'privateKey' in newCredentials.credentials &&
      typeof newCredentials.credentials.privateKey === 'string'
    ) {
      newCredentials.credentials.privateKey = window.btoa(
        newCredentials.credentials.privateKey
      );
    }
    updateCredentialMutation.mutate({
      credential: newCredentials,
    });
    router.stateService.go('portainer.settings.sharedcredentials');
  }
}

function getFormType(credentialType: CredentialType) {
  switch (credentialType) {
    case CredentialType.GOOGLE_CLOUD:
      return FormTypes.SERVICE_ACCOUNT;

    case CredentialType.AWS:
      return FormTypes.ACCESS_KEY;

    case CredentialType.AZURE:
      return FormTypes.AZURE;

    case CredentialType.SSH:
      return FormTypes.USERNAME_PASSWORD;

    case CredentialType.CIVO:
    case CredentialType.DIGITAL_OCEAN:
    case CredentialType.LINODE:
    default:
      return FormTypes.API;
  }
}
