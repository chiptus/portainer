import { EdgeGroup } from '@/react/edge/edge-groups/types';

export enum EdgeConfigurationTypeString {
  General = 'general',
  DeviceSpecificFile = 'filename',
  DeviceSpecificFolder = 'foldername',
}

export enum EdgeConfigurationCategoryString {
  Configuration = 'configuration',
  Secret = 'secret',
}

type EdgeConfigurationPayloadField = {
  name: string;
  baseDir?: string;
  edgeGroupIDs: EdgeGroup['Id'][];
  type: EdgeConfigurationTypeString;
  category: EdgeConfigurationCategoryString;
};

type WebEditor = {
  files: {
    name: string;
    content: string;
  }[];
  file?: never;
};

type Archive = {
  file: File;
  files?: never;
};

export type EdgeConfigurationCreatePayload = {
  edgeConfiguration: EdgeConfigurationPayloadField;
} & (Archive | WebEditor);
