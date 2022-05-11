import { object, string } from 'yup';

export function noDuplicateNamesSchema(names: string[]) {
  return string()
    .required('Name is required')
    .test('not existing name', 'Name is already in use', (newName) =>
      names.every((name) => name !== newName)
    );
}

export function validationSchema(names: string[]) {
  return object().shape({
    name: noDuplicateNamesSchema(names),
    credentials: object()
      .shape({
        apiKey: string().required('API key is required'),
      })
      .required(),
  });
}
