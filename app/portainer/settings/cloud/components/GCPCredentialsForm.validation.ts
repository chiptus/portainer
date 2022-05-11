import { object, string } from 'yup';

import { noDuplicateNamesSchema } from './APICredentialsForm.validation';

export function validationSchema(names: string[]) {
  return object().shape({
    name: noDuplicateNamesSchema(names),
    credentials: object()
      .shape({
        jsonKeyBase64: string().required('Service account key is required'),
      })
      .required(),
  });
}
