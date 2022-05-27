import axios, { parseAxiosError } from '@/portainer/services/axios';
import { KaasProvider, Credential } from '@/portainer/settings/cloud/types';

import { Environment } from '../types';

import {
  CreateApiClusterPayload,
  CreateGKEClusterPayload,
  KaasInfoResponse,
} from './types';
import { parseKaasInfoResponse } from './converter';

function buildUrl(provider: KaasProvider, action: string) {
  return `/cloud/${provider}/${action}`;
}

export async function createKaasEnvironment(
  payload: CreateApiClusterPayload | CreateGKEClusterPayload,
  provider: KaasProvider
) {
  try {
    const { data } = await axios.post<Environment>(
      buildUrl(provider, 'cluster'),
      payload
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

export async function getKaasInfo(
  provider: KaasProvider,
  credential: Credential
) {
  if (provider === credential.provider) {
    try {
      const { data } = await axios.get<KaasInfoResponse>(
        buildUrl(provider, 'info'),
        { params: { credentialId: credential.id } }
      );
      return parseKaasInfoResponse(data);
    } catch (e) {
      throw parseAxiosError(e as Error);
    }
  }
  return null;
}
