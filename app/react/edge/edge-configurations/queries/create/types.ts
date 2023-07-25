import { EdgeGroup } from '@/react/edge/edge-groups/types';

export enum EdgeConfigurationTypeString {
  General = 'general',
  DeviceSpecificFile = 'filename',
  DeviceSpecificFolder = 'foldername',
}

type EdgeConfigurationPayloadField = {
  name: string;
  baseDir?: string;
  edgeGroupIDs: EdgeGroup['Id'][];
  type: EdgeConfigurationTypeString;
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
