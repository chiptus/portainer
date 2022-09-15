import { object, string } from 'yup';

export function noDuplicateNamesSchema(names: string[]) {
  return string()
    .required('Name is required')
    .test('not existing name', 'Name is already in use', (newName) =>
      names.every((name) => name !== newName)
    );
}

export function validationSchema(names: string[], isEditing = false) {
  return object().shape({
    name: noDuplicateNamesSchema(names)
      .matches(/^[-_a-z0-9]+$/, {
        message:
          "This field must consist of lower case alphanumeric characters, '_' or '-' (e.g. 'my-name', or 'abc-123').",
      })
      .required(),
    username: string().optional(),
    password: isEditing
      ? string().notRequired()
      : string().required('personal access token is required'),
  });
}
