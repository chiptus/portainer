export interface CloudSettingsFormValues {
  CivoApiKey?: string;
  LinodeToken?: string;
  DigitalOceanToken?: string;
}

export interface CloudFeatureSettings {
  linode?: boolean;
  digitalOcean?: boolean;
}

export type ApiKeyTypes = keyof CloudSettingsFormValues;

export interface CloudSettingsAPIPayload {
  CloudApiKeys: {
    CivoApiKey?: string;
    DigitalOceanToken?: string;
    LinodeToken?: string;
  };
}
