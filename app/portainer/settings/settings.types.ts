import { CloudApiKeys } from '../environments/components/kaas/kaas.types';

import {
  CloudSettingsAPIPayload,
  CloudFeatureSettings,
} from './cloud/cloud.types';

enum AuthenticationMethod {
  // AuthenticationInternal represents the internal authentication method (authentication against Portainer API)
  AuthenticationInternal,
  // AuthenticationLDAP represents the LDAP authentication method (authentication against a LDAP server)
  AuthenticationLDAP,
  // AuthenticationOAuth represents the OAuth authentication method (authentication against a authorization server)
  AuthenticationOAuth,
}

export interface SettingsResponse {
  LogoURL: string;
  BlackListedLabels: { name: string; value: string }[];
  AuthenticationMethod: AuthenticationMethod;
  SnapshotInterval: string;
  TemplatesURL: string;
  EdgeAgentCheckinInterval: number;
  EnableEdgeComputeFeatures: boolean;
  UserSessionTimeout: string;
  KubeconfigExpiry: string;
  EnableTelemetry: boolean;
  HelmRepositoryURL: string;
  KubectlShellImage: string;
  DisableTrustOnFirstConnect: boolean;
  EnforceEdgeID: boolean;
  AgentSecret: string;
  CloudApiKeys: CloudApiKeys;
}

// expand this type with sub-modules payloads interfaces
// so each sub-module can use settings.service.updateSettings()
// with it own sub-payload
// see cloud.service.ts and settings.service.ts
export type SettingsAPIPayload = CloudSettingsAPIPayload;
export type SettingsCloudFeature = CloudFeatureSettings;
