import { object, string } from 'yup';

import { noDuplicateNamesSchema } from './APICredentialsForm.validation';

export function validationSchema(names: string[], isEditing = false) {
  if (isEditing) {
    return object().shape({
      name: noDuplicateNamesSchema(names),
      credentials: object()
        .shape({
          accessKeyId: string().required('Access key id is required'),
          secretAccessKey: string(),
        })
        .required(),
    });
  }
  return object().shape({
    name: noDuplicateNamesSchema(names),
    credentials: object()
      .shape({
        accessKeyId: string().required('Access key id is required'),
        secretAccessKey: string().required('Secret access key is required'),
      })
      .required(),
  });
}
