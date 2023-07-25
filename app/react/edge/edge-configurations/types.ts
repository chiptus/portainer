import { User } from '@/portainer/users/types';
import { EdgeGroup } from '@/react/edge/edge-groups/types';

// extracted from iota values
export enum EdgeConfigurationType {
  EdgeConfigTypeGeneral = 1,
  EdgeConfigTypeSpecificFile,
  EdgeConfigTypeSpecificFolder,
}

type EdgeConfigurationProgress = {
  success: number;
  total: number;
};

export type EdgeConfiguration = {
  id: number;
  name: string;
  type: EdgeConfigurationType;
  // state:        EdgeConfigStateType ;
  edgeGroupIDs: EdgeGroup['Id'][];
  baseDir: string;
  created: number;
  createdBy: User['Id'];
  updated?: number;
  updatedBy?: User['Id'];
  progress: EdgeConfigurationProgress;
};
