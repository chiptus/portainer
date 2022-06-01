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

export const sensitiveFields = [
  'jsonKeyBase64',
  'apiKey',
  'secretAccessKey',
  'clientSecret',
];

export function sensitiveFieldChanged(credentials: CredentialDetails) {
  const newCredentialsSensitive = _.pick(credentials, sensitiveFields);

  return Object.values(newCredentialsSensitive).some((sensitiveValue) => {
    if (sensitiveValue.trim().length === 0) {
      return false;
    }
    return true;
  });
}
