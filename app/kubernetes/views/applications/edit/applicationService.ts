import axios, { parseAxiosError } from '@/portainer/services/axios';

export async function rolloutRestartApplication(
  environmentId: number,
  namespace: string,
  kind: string,
  name: string
) {
  try {
    await axios.get(
      `/kubernetes/${environmentId}/namespaces/${namespace}/applications/${kind}/${name}?rollout-restart=true`
    );
  } catch (error) {
    throw parseAxiosError(
      error as Error,
      `Failed to restart application ${kind}/${name} in namespace ${namespace}`
    );
  }
}
