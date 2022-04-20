import { boolean, number, object, string } from 'yup';

export function validationSchema() {
  return object().shape({
    name: string().required('Name is required'),
    token: string(),
    portainerUrl: string().required('Portainer URL is required'),
    pollFrequency: number().required(),
    allowSelfSignedCertificates: boolean(),
    authEnabled: boolean(),
    envVars: string(),
  });
}
