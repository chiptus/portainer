import {
  PaginationTableSettings,
  SortableTableSettings,
} from '@/portainer/components/datatables/types';

export interface CredentialTableSettings
  extends SortableTableSettings,
    PaginationTableSettings {}

export enum KaasProvider {
  CIVO = 'civo',
  LINODE = 'linode',
  DIGITAL_OCEAN = 'digitalocean',
  GOOGLE_CLOUD = 'googlecloud',
  AWS = 'aws',
  AZURE = 'azure',
}

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
