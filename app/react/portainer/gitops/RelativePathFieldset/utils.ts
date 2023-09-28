import { EdgeStack } from '@/react/edge/edge-stacks/types';

import { RelativePathModel } from '../types';

export function parseRelativePathResponse(stack: EdgeStack): RelativePathModel {
  return {
    SupportRelativePath: stack.SupportRelativePath,
    FilesystemPath: stack.FilesystemPath,
    SupportPerDeviceConfigs: stack.SupportPerDeviceConfigs,
    PerDeviceConfigsMatchType: stack.PerDeviceConfigsMatchType,
    PerDeviceConfigsGroupMatchType: stack.PerDeviceConfigsGroupMatchType,
    PerDeviceConfigsPath: stack.PerDeviceConfigsPath,
  };
}
