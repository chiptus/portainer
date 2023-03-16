import {
  K8sDistributionType,
  KaasProvider,
  ProvisionOption,
} from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/types';

export enum CredentialType {
  CIVO = 'civo',
  LINODE = 'linode',
  DIGITAL_OCEAN = 'digitalocean',
  GOOGLE_CLOUD = 'gke',
  AWS = 'amazon',
  AZURE = 'azure',
  SSH = 'ssh',
}

export const credentialTitles: Record<CredentialType, string> = {
  civo: 'Civo',
  linode: 'Linode',
  digitalocean: 'DigitalOcean',
  gke: 'Google Cloud',
  amazon: 'AWS',
  azure: 'Azure',
  ssh: 'SSH',
};

// ssh credentials will eventually have more than one valid distribution type
// create a mapping of credential types to distribution/provider types
export const credentialTypeToProvidersMap: Record<
  CredentialType,
  ProvisionOption[]
> = {
  ssh: [K8sDistributionType.MICROK8S],
  civo: [KaasProvider.CIVO],
  linode: [KaasProvider.LINODE],
  digitalocean: [KaasProvider.DIGITAL_OCEAN],
  gke: [KaasProvider.GOOGLE_CLOUD],
  amazon: [KaasProvider.AWS],
  azure: [KaasProvider.AZURE],
};

export const providerToCredentialTypeMap: Record<
  ProvisionOption,
  CredentialType
> = {
  [K8sDistributionType.MICROK8S]: CredentialType.SSH,
  [KaasProvider.CIVO]: CredentialType.CIVO,
  [KaasProvider.LINODE]: CredentialType.LINODE,
  [KaasProvider.DIGITAL_OCEAN]: CredentialType.DIGITAL_OCEAN,
  [KaasProvider.GOOGLE_CLOUD]: CredentialType.GOOGLE_CLOUD,
  [KaasProvider.AWS]: CredentialType.AWS,
  [KaasProvider.AZURE]: CredentialType.AZURE,
};

export const credentialTypeHelpLinks: Record<CredentialType, string> = {
  civo: 'https://docs.portainer.io/admin/settings/credentials/civo',
  linode: 'https://docs.portainer.io/admin/settings/credentials/linode',
  digitalocean:
    'https://docs.portainer.io/admin/settings/credentials/digitalocean',
  gke: 'https://docs.portainer.io/admin/settings/credentials/gke',
  amazon: 'https://docs.portainer.io/admin/settings/credentials/eks',
  azure: 'https://docs.portainer.io/admin/settings/credentials/aks',
  ssh: 'https://docs.portainer.io/admin/settings/credentials/ssh',
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

export type SSHCredentials = {
  username: string;
  password: string;
  privateKey: string;
  passphrase: string;
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

export interface SSHCredentialFormValues {
  name: string;
  credentials: SSHCredentials;
}

export type CredentialDetails =
  | APICredentials
  | AccessKeyCredentials
  | ServiceAccountCredentials
  | AzureCredentials
  | SSHCredentials;

export type GenericFormValues = {
  name: string;
  credentials: CredentialDetails;
};

export interface CreateCredentialPayload {
  name: string;
  provider: CredentialType;
  credentials: CredentialDetails;
}

export interface UpdateCredentialPayload {
  name?: string;
  provider?: CredentialType;
  credentials?: Partial<CredentialDetails>;
}

// using a type instead of interface to conform to the Record<string, unknown> types required in the datatable component
export type Credential = {
  id: number;
  created: number;
  name: string;
  provider: CredentialType;
  credentials: CredentialDetails;
};
