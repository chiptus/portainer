import { react2angular } from '@/react-tools/react2angular';
import { EditorType } from '@/edge/types';

import { BoxSelector } from '@@/BoxSelector';
import { BoxSelectorOption } from '@@/BoxSelector/types';

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
      id: 'deployment_compose',
      icon: 'fab fa-docker',
      label: 'Compose',
      description: 'Docker compose format',
      value: EditorType.Compose,
      disabled: () => hasNomadEndpoint,
      tooltip: () =>
        hasNomadEndpoint
          ? 'Cannot use this option with Edge Nomad endpoints'
          : '',
    },
    {
      id: 'deployment_kube',
      icon: 'fa fa-cubes',
      label: 'Kubernetes',
      description: 'Kubernetes manifest format',
      value: EditorType.Kubernetes,
      disabled: () => hasDockerEndpoint || hasNomadEndpoint,
      tooltip: () =>
        hasDockerEndpoint || hasNomadEndpoint
          ? 'Cannot use this option with Edge Docker or Edge Nomad endpoints'
          : '',
    },
    {
      id: 'deployment_nomad',
      icon: 'nomad-icon',
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
