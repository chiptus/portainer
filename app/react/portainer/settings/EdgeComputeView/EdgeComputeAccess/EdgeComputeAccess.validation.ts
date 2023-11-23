import { array, number, object, string } from 'yup';

export function validationSchema() {
  return object().shape({
    selectedUsersAndTeams: array(
      object().shape({
        Type: string().oneOf(['team', 'user']),
        Name: string(),
        Id: number(),
      })
    ).min(1),
  });
}
