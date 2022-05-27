import { object } from 'yup';

import { nameValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';

export function validationSchema() {
  return object().shape({
    name: nameValidation()
      .matches(
        /^[a-z0-9-]+$/,
        'Name must only contain lowercase alphanumeric characters and hyphens.'
      )
      .max(32, 'Name must be 32 characters or less.'),
  });
}
