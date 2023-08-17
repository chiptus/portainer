import * as JsonPatch from 'fast-json-patch';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';

const apiKindMap = {
  Service: 'services',
  Deployment: 'deployments',
  Namespace: 'namespaces',
  Secret: 'secrets',
  ConfigMap: 'configmaps',
  PersistentVolumeClaim: 'persistentvolumeclaims',
} as const;

const apiVersionMap = {
  v1: 'api/v1',
  'apps/v1': 'apis/apps/v1',
} as const;

export async function applyPatch(
  kind: keyof typeof apiKindMap,
  apiVersion: keyof typeof apiVersionMap,
  environmentId: EnvironmentId,
  namespace: string,
  name: string,
  patch: JsonPatch.Operation[]
) {
  try {
    const { data: deployment } = await axios.patch(
      buildUrl(kind, apiVersion, environmentId, namespace, name),
      patch,
      {
        headers: {
          'Content-Type': 'application/json-patch+json',
        },
      }
    );
    return deployment;
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to update YAML resource');
  }
}

function buildUrl(
  kind: keyof typeof apiKindMap,
  apiVersion: keyof typeof apiVersionMap,
  environmentId: EnvironmentId,
  namespace: string,
  name: string
) {
  let resourcePrefix: string = apiKindMap[kind];
  if (!resourcePrefix) {
    resourcePrefix = `${kind.toLowerCase()}s`; // pluralize the resource kind; DOES not support all the edge cases
  }

  const apiVersionURI = apiVersionMap[apiVersion];

  let url = `/endpoints/${environmentId}/kubernetes/${apiVersionURI}`;

  if (namespace) {
    url += `/namespaces/${namespace}`;
  }

  if (resourcePrefix && name) {
    url += `/${resourcePrefix}/${name}`;
  }

  url += `?fieldManager=kubectl`;

  return url;
}
