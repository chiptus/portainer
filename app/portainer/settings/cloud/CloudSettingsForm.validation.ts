import { object, string } from 'yup';

export function validationSchema() {
  return object().shape({
    civo: string().notRequired(),
    linode: string().notRequired(),
    digitalOcean: string().notRequired(),
  });
}
