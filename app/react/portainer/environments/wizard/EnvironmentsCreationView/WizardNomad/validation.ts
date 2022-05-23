import { boolean, number, object, SchemaOf, string } from 'yup';

import { metadataValidation } from '../shared/MetadataFieldset/validation';
import { nameValidation } from '../shared/NameField';

import { FormValues } from './types';

export function validationSchema(): SchemaOf<FormValues> {
  return object().shape({
    name: nameValidation(),
    token: string().default(''),
    portainerUrl: string().required('Portainer URL is required'),
    pollFrequency: number().required(),
    allowSelfSignedCertificates: boolean().default(true),
    authEnabled: boolean().default(true),
    envVars: string().default(''),
    meta: metadataValidation(),
  });
}
