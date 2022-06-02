import { object } from 'yup';

import { nameValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';

export function validationSchema() {
  return object().shape({
    name: nameValidation()
      .matches(
        /^[a-z0-9-]+$/,
        'Name must only contain lowercase alphanumeric characters and hyphens.'
      )
      .matches(
        /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/,
        'Name must start and end with an alphanumeric character.'
      )
      .max(32, 'Name must be 32 characters or less.'),
  });
}
