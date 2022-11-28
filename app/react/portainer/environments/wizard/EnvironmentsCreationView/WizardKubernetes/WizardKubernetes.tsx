import { useState } from 'react';
import { Zap, Cloud, UploadCloud } from 'lucide-react';

import {
  Environment,
  EnvironmentCreationTypes,
} from '@/react/portainer/environments/types';
import { commandsTabs } from '@/react/edge/components/EdgeScriptForm/scripts';

import { BoxSelector } from '@@/BoxSelector';
import { BoxSelectorOption } from '@@/BoxSelector/types';
import { BadgeIcon } from '@@/BadgeIcon';

import { AnalyticsStateKey } from '../types';
import { EdgeAgentTab } from '../shared/EdgeAgentTab';
import { useFilterEdgeOptionsIfNeeded } from '../useOnlyEdgeOptions';

import { AgentPanel } from './AgentPanel';
import { KubeConfigForm } from './KubeConfig/index';

interface Props {
  onCreate(environment: Environment, analytics: AnalyticsStateKey): void;
}

const defaultOptions: BoxSelectorOption<EnvironmentCreationTypes>[] = [
  {
    id: 'agent_endpoint',
    icon: <BadgeIcon icon={Zap} size="3xl" />,
    label: 'Agent',
    value: EnvironmentCreationTypes.AgentEnvironment,
    description: '',
  },
  {
    id: 'edgeAgent',
    icon: <BadgeIcon icon={Cloud} size="3xl" />,
    label: 'Edge Agent',
    description: '',
    value: EnvironmentCreationTypes.EdgeAgentEnvironment,
  },
  {
    id: 'kubeconfig_endpoint',
    icon: <BadgeIcon icon={UploadCloud} size="3xl" />,
    label: 'Import',
    value: EnvironmentCreationTypes.KubeConfigEnvironment,
    description: 'Import an existing Kubernetes config',
  },
];

export function WizardKubernetes({ onCreate }: Props) {
  const options = useFilterEdgeOptionsIfNeeded(
    defaultOptions,
    EnvironmentCreationTypes.EdgeAgentEnvironment
  );

  const [creationType, setCreationType] = useState(options[0].value);

  const tab = getTab(creationType);

  return (
    <div className="form-horizontal">
      <BoxSelector
        onChange={(v) => setCreationType(v)}
        options={options}
        value={creationType}
        radioName="creation-type"
      />

      {tab}
    </div>
  );

  function getTab(type: typeof options[number]['value']) {
    switch (type) {
      case EnvironmentCreationTypes.AgentEnvironment:
        return (
          <AgentPanel
            onCreate={(environment) => onCreate(environment, 'kubernetesAgent')}
          />
        );
      case EnvironmentCreationTypes.EdgeAgentEnvironment:
        return (
          <EdgeAgentTab
            onCreate={(environment) =>
              onCreate(environment, 'kubernetesEdgeAgent')
            }
            commands={[{ ...commandsTabs.k8sLinux, label: 'Linux' }]}
          />
        );
      case EnvironmentCreationTypes.KubeConfigEnvironment:
        return (
          <KubeConfigForm
            onCreate={(environment) => onCreate(environment, 'kubernetesAgent')}
          />
        );
      default:
        throw new Error('Creation type not supported');
    }
  }
}
