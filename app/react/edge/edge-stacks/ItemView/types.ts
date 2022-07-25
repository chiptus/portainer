import { EnvironmentId } from '@/portainer/environments/types';

export interface StackEnvironmentViewModel {
  id: EnvironmentId;
  name: string;
  asyncMode: boolean;
  status: {
    Type: number;
    Error: string;
  };
  hasLogs: boolean;
}
