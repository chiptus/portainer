import _ from 'lodash';

import { EnvironmentType } from '@/react/portainer/environments/types';

import { EditorType } from './types';

export function getValidEditorTypes(
  endpointTypes: EnvironmentType[],
  allowKubeToSelectCompose?: boolean
) {
  const right: Partial<Record<EnvironmentType, EditorType[]>> = {
    [EnvironmentType.EdgeAgentOnDocker]: [EditorType.Compose],
    // allow kube to view compose if it's an existing kube compose stack
    [EnvironmentType.EdgeAgentOnKubernetes]: allowKubeToSelectCompose
      ? [EditorType.Kubernetes, EditorType.Compose]
      : [EditorType.Kubernetes],
    [EnvironmentType.EdgeAgentOnNomad]: [EditorType.Nomad],
  };

  return endpointTypes.length
    ? _.intersection(...endpointTypes.map((type) => right[type]))
    : [EditorType.Compose, EditorType.Kubernetes, EditorType.Nomad];
}
