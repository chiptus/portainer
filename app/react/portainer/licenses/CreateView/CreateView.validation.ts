import { object, string } from 'yup';

export function validationSchema() {
  return object().shape({
    key: string().required('License key is required.'),
  });
}
