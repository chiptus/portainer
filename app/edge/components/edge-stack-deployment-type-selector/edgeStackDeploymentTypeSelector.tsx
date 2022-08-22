import { react2angular } from '@/react-tools/react2angular';
import { EditorType } from '@/edge/types';
import NomadIcon from '@/assets/ico/vendor/nomad.svg?c';

import { BoxSelector } from '@@/BoxSelector';
import { BoxSelectorOption } from '@@/BoxSelector/types';
import {
  compose,
  kubernetes,
} from '@@/BoxSelector/common-options/deployment-methods';

interface Props {
  value: number;
  onChange(value: number): void;
  hasDockerEndpoint: boolean;
  hasKubeEndpoint: boolean;
  hasNomadEndpoint: boolean;
}

export function EdgeStackDeploymentTypeSelector({
  value,
  onChange,
  hasDockerEndpoint,
  hasKubeEndpoint,
  hasNomadEndpoint,
}: Props) {
  const deploymentOptions: BoxSelectorOption<number>[] = [
    {
      ...compose,
      value: EditorType.Compose,
      disabled: () => hasNomadEndpoint,
      tooltip: () =>
        hasNomadEndpoint
          ? 'Cannot use this option with Edge Nomad endpoints'
          : '',
    },
    {
      ...kubernetes,
      value: EditorType.Kubernetes,
      disabled: () => hasDockerEndpoint || hasNomadEndpoint,
      tooltip: () =>
        hasDockerEndpoint || hasNomadEndpoint
          ? 'Cannot use this option with Edge Docker or Edge Nomad endpoints'
          : '',
    },
    {
      id: 'deployment_nomad',
      icon: NomadIcon,
      label: 'Nomad',
      description: 'Nomad HCL format',
      value: EditorType.Nomad,
      disabled: () => hasDockerEndpoint || hasKubeEndpoint,
      tooltip: () =>
        hasDockerEndpoint || hasKubeEndpoint
          ? 'Cannot use this option with Edge Docker or Edge Kubernetes endpoints'
          : '',
    },
  ];

  return (
    <>
      <div className="col-sm-12 form-section-title"> Deployment type</div>
      <BoxSelector
        radioName="deploymentType"
        value={value}
        options={deploymentOptions}
        onChange={onChange}
      />
    </>
  );
}

export const EdgeSTackDeploymentTypeSelectorAngular = react2angular(
  EdgeStackDeploymentTypeSelector,
  [
    'value',
    'onChange',
    'hasDockerEndpoint',
    'hasKubeEndpoint',
    'hasNomadEndpoint',
  ]
);
