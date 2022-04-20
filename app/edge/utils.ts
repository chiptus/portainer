import _ from 'lodash';

import { EnvironmentType } from '@/portainer/environments/types';

import { EditorType } from './types';

export function getValidEditorTypes(endpointTypes: EnvironmentType[]) {
  const right: Partial<Record<EnvironmentType, EditorType[]>> = {
    [EnvironmentType.EdgeAgentOnDocker]: [EditorType.Compose],
    [EnvironmentType.EdgeAgentOnKubernetes]: [
      EditorType.Compose,
      EditorType.Kubernetes,
    ],
    [EnvironmentType.EdgeAgentOnNomad]: [EditorType.Nomad],
  };

  return endpointTypes.length
    ? _.intersection(...endpointTypes.map((type) => right[type]))
    : [EditorType.Compose, EditorType.Kubernetes, EditorType.Nomad];
}
