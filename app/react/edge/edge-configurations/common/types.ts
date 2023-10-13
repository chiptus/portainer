import { EdgeGroup } from '@/react/edge/edge-groups/types';
import { EdgeConfigurationCategoryString } from '@/react/edge/edge-configurations/queries/create/types';

export enum FormValuesFileMethod {
  File = 'file',
  Archive = 'archive',
}

export enum FormValuesEdgeConfigurationType {
  General = 'general',
  DeviceSpecific = 'device-specific',
}

export enum FormValuesEdgeConfigurationMatchingRule {
  MatchFolder = 'foldername',
  MatchFile = 'filename',
}

export type FormValues = {
  name: string;
  groupIds: EdgeGroup['Id'][];
  directory: string;
  type: FormValuesEdgeConfigurationType;
  category: EdgeConfigurationCategoryString;
  matchingRule?: FormValuesEdgeConfigurationMatchingRule;
  file:
    | {
        // default value (empty)
        name: string;
        method?: never;
        content?: never;
      }
    | {
        // upload from archive
        name: string;
        method: FormValuesFileMethod.Archive;
        content: File;
      }
    | {
        // files from webeditor
        name: string;
        method: FormValuesFileMethod.File;
        content: string;
      };
};
