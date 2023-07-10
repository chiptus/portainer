import { EndpointSettings } from 'docker-types/generated/1.41';
import { AxiosRequestHeaders } from 'axios';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { buildUrl } from './buildUrl';

interface ConnectContainerPayload {
  Container: string;
  EndpointConfig?: EndpointSettings;
}

export async function connectContainer({
  environmentId,
  containerId,
  networkId,
  aliases,
  nodeName,
}: {
  environmentId: EnvironmentId;
  networkId: string;
  containerId: string;
  aliases?: EndpointSettings['Aliases'];
  nodeName?: string;
}) {
  const payload: ConnectContainerPayload = {
    Container: containerId,
  };
  if (aliases) {
    payload.EndpointConfig = {
      Aliases: aliases,
    };
  }

  const headers: AxiosRequestHeaders = {};

  if (nodeName) {
    headers['X-PortainerAgent-Target'] = nodeName;
  }

  try {
    await axios.post(
      buildUrl(environmentId, { id: networkId, action: 'connect' }),
      payload
    );
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to connect container');
  }
}
