import { trimSHA } from '@/docker/filters/utils';

import { DockerImage } from './types';
import { DockerImageResponse } from './types/response';

export function parseViewModel(response: DockerImageResponse): DockerImage {
  return {
    ...response,
    Used: false,
    RepoTags:
      response.RepoTags ??
      response.RepoDigests.map((digest) => `${trimSHA(digest)}:<none>`),
  };
}
