import _ from 'lodash';

import { CredentialDetails } from './types';

// trimObject trims the string values in an object (shallow) and returns the trimmed object
export function trimObject<T extends Record<string, unknown>>(obj: T) {
  return _.mapValues(obj, (value) => {
    if (typeof value === 'string') {
      return value.trim();
    }
    return value;
  });
}

type KeysOfUnion<T> = T extends T ? keyof T : never;

export const sensitiveFields: KeysOfUnion<CredentialDetails>[] = [
  'jsonKeyBase64',
  'apiKey',
  'secretAccessKey',
  'clientSecret',
  'password',
  'privateKey',
  'passphrase',
];

export function getUnchangedSensitiveFields(credentials: CredentialDetails) {
  const newCredentialsSensitive = _.pick(credentials, sensitiveFields);

  const sensitiveKeysThatHaveChanged = Object.entries(newCredentialsSensitive)
    .filter(([, value]) => value.trim().length === 0)
    .map(([key]) => key);

  return sensitiveKeysThatHaveChanged;
}
