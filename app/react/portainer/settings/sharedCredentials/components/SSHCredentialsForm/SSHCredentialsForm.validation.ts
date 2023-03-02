import { object, string } from 'yup';

import { noDuplicateNamesSchema } from '../APICredentialsForm.validation';

export function validationSchema(names: string[]) {
  return object().shape({
    name: noDuplicateNamesSchema(names),
    credentials: object()
      .shape({
        username: string().required('Username is required.'),
        password: string(),
        privateKey: string().when('passphrase', {
          is: (passphrase: string) => !!passphrase,
          then: string().required(
            'SSH private key is required when a passphrase is set.'
          ),
          otherwise: string(),
        }),
        passphrase: string(),
      })
      .required(),
  });
}
