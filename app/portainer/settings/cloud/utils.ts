import _ from 'lodash';

import { KaasProvider, GenericFormValues, CredentialDetails } from './types';

const providerTitles = {
  civo: 'Civo',
  linode: 'Linode',
  digitalocean: 'DigitalOcean',
  googlecloud: 'Google Cloud',
  aws: 'AWS',
  azure: 'Azure',
};

export function getProviderTitle(provider: KaasProvider) {
  return providerTitles[provider];
}

// trimObject trims the string values in an object (shallow) and returns the trimmed object
export function trimObject<T extends Record<string, unknown>>(obj: T) {
  return _.mapValues(obj, (value) => {
    if (typeof value === 'string') {
      return value.trim();
    }
    return value;
  });
}

export const sensitiveFields = [
  'jsonKeyBase64',
  'apiKey',
  'secretAccessKey',
  'clientSecret',
];

// isMeaningfulChange only returns true if the credential form values have meaningfully changed
export function isMeaningfulChange<T extends Partial<GenericFormValues>>(
  newFormValues: T,
  oldFormValues: T
) {
  // if any value other than the sensitive credential is changed, then return true
  const newCredentialsNotSensitive = _.omit(
    newFormValues.credentials,
    sensitiveFields
  );
  const oldCredentialsNotSensitive = _.omit(
    oldFormValues.credentials,
    sensitiveFields
  );
  const newFormNotSensitive = {
    ...newFormValues,
    credentials: newCredentialsNotSensitive,
  };
  const oldFormNotSensitive = {
    ...oldFormValues,
    credentials: oldCredentialsNotSensitive,
  };
  if (!_.isEqual(newFormNotSensitive, oldFormNotSensitive)) {
    return true;
  }

  if (newFormValues.credentials) {
    return sensitiveFieldChanged(newFormValues.credentials);
  }
  return false;
}

export function sensitiveFieldChanged(credentials: CredentialDetails) {
  const newCredentialsSensitive = _.pick(credentials, sensitiveFields);

  return Object.values(newCredentialsSensitive).some((sensitiveValue) => {
    // if all characters are '*', then return false
    if (/^[*]+$/.test(sensitiveValue.trim())) {
      return false;
    }
    return true;
  });
}
