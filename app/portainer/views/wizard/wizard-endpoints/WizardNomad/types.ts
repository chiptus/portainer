export interface EdgeInfo {
  id?: string;
  key?: string;
}

export interface FormValues {
  name: string;
  token: string;
  portainerUrl: string;
  pollFrequency: number;
  allowSelfSignedCertificates: boolean;
  authEnabled: boolean;
  envVars: string;
}
