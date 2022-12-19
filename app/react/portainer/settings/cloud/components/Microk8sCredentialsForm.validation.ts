import { object, string } from 'yup';

import { noDuplicateNamesSchema } from './APICredentialsForm.validation';

export function validationSchema(names: string[], isEditing = false) {
  if (isEditing) {
    return object().shape({
      name: noDuplicateNamesSchema(names),
      credentials: object()
        .shape({
          username: string().required('Username is required'),
          password: string(),
        })
        .required(),
    });
  }
  return object().shape({
    name: noDuplicateNamesSchema(names),
    credentials: object()
      .shape({
        username: string().required('Username is required'),
        password: string().required('Password is required'),
      })
      .required(),
  });
}
