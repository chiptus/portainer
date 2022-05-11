import { object, string } from 'yup';

import { noDuplicateNamesSchema } from './APICredentialsForm.validation';

export function validationSchema(names: string[]) {
  return object().shape({
    name: noDuplicateNamesSchema(names),
    credentials: object()
      .shape({
        clientID: string().required('Client ID is required'),
        clientSecret: string().required('Client Secret is required'),
        tenantID: string().required('Tenant ID key is required'),
        subscriptionID: string().required('Subscription ID is required'),
      })
      .required(),
  });
}
