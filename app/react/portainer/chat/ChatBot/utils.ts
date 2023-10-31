import { EnvironmentType, PlatformType } from '../../environments/types';
import { getPlatformType } from '../../environments/utils';

export function getPlatformTypeForAnalytics(environmentType?: EnvironmentType) {
  switch (getPlatformType(environmentType)) {
    case PlatformType.Kubernetes:
      return 'kubernetes';
    case PlatformType.Docker:
      return 'docker';
    case PlatformType.Azure:
      return 'azure';
    default:
      return 'unknown';
  }
}
