import { object, number, array, SchemaOf, string } from 'yup';

import { EnvironmentMetadata } from '@/react/portainer/environments/environment.service/create';

export function metadataValidation(): SchemaOf<EnvironmentMetadata> {
  return object({
    groupId: number(),
    tagIds: array().of(number()).default([]),
    customTemplateId: number().default(0),
    variables: object().default({}),
    customTemplateContent: string().default(''),
  });
}
