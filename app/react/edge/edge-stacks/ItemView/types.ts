import { EnvironmentId } from '@/react/portainer/environments/types';

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
