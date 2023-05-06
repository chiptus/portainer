import { EnvironmentId } from '@/react/portainer/environments/types';
import axios from '@/portainer/services/axios';

interface CreateStackCommandRequest {
  StackOperation: string;
}

export async function removeStack(endpointId: EnvironmentId, stackId: number) {
  const payload: CreateStackCommandRequest = {
    StackOperation: 'remove',
  };
  await axios.post<void>(urlBuilder(endpointId, stackId), payload);
}

export function urlBuilder(endpointId: EnvironmentId, stackId: number) {
  return `/endpoints/${endpointId}/edge/async/commands/stack/${stackId}`;
}
