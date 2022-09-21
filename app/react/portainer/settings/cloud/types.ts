import {
  PaginationTableSettings,
  SortableTableSettings,
} from '@@/datatables/types-old';

export interface CredentialTableSettings
  extends SortableTableSettings,
    PaginationTableSettings {}

export enum KaasProvider {
  CIVO = 'civo',
  LINODE = 'linode',
  DIGITAL_OCEAN = 'digitalocean',
  GOOGLE_CLOUD = 'gke',
  AWS = 'amazon',
  AZURE = 'azure',
}

export const providerTitles: Record<KaasProvider, string> = {
  civo: 'Civo',
  linode: 'Linode',
  digitalocean: 'DigitalOcean',
  gke: 'Google Cloud',
  amazon: 'AWS',
  azure: 'Azure',
};

export const providerHelpLinks: Record<KaasProvider, string> = {
  civo: 'https://docs.portainer.io/admin/settings/cloud/civo',
  linode: 'https://docs.portainer.io/admin/settings/cloud/linode',
  digitalocean: 'https://docs.portainer.io/admin/settings/cloud/digitalocean',
  gke: 'https://docs.portainer.io/admin/settings/cloud/gke',
  amazon: 'https://docs.portainer.io/admin/settings/cloud/eks',
  azure: 'https://docs.portainer.io/admin/settings/cloud/aks',
};

export type APICredentials = {
  apiKey: string;
};

export type AccessKeyCredentials = {
  accessKeyId: string;
  secretAccessKey: string;
};

export type ServiceAccountCredentials = {
  jsonKeyBase64: string;
};

export type AzureCredentials = {
  clientID: string;
  clientSecret: string;
  tenantID: string;
  subscriptionID: string;
};

export interface APIFormValues {
  name: string;
  credentials: APICredentials;
}

export interface AccessKeyFormValues {
  name: string;
  credentials: AccessKeyCredentials;
}

export interface ServiceAccountFormValues {
  name: string;
  credentials: ServiceAccountCredentials;
}

export interface AzureFormValues {
  name: string;
  credentials: AzureCredentials;
}

export type CredentialDetails =
  | APICredentials
  | AccessKeyCredentials
  | ServiceAccountCredentials
  | AzureCredentials;

export type GenericFormValues = {
  name: string;
  credentials: CredentialDetails;
};

export interface CreateCredentialPayload {
  name: string;
  provider: KaasProvider;
  credentials: CredentialDetails;
}

export interface UpdateCredentialPayload {
  name?: string;
  provider?: KaasProvider;
  credentials?: Partial<CredentialDetails>;
}

// using a type instead of interface to conform to the Record<string, unknown> types required in the datatable component
export type Credential = {
  id: number;
  created: number;
  name: string;
  provider: KaasProvider;
  credentials: CredentialDetails;
};
