import { useState } from 'react';
import { Zap, UploadCloud } from 'lucide-react';
import _ from 'lodash';

import { Environment } from '@/react/portainer/environments/types';
import { commandsTabs } from '@/react/edge/components/EdgeScriptForm/scripts';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import EdgeAgentStandardIcon from '@/react/edge/components/edge-agent-standard.svg?c';
import EdgeAgentAsyncIcon from '@/react/edge/components/edge-agent-async.svg?c';
import {
  CustomTemplate,
  CustomTemplateKubernetesType,
} from '@/react/portainer/custom-templates/types';
import { useCustomTemplates } from '@/react/portainer/custom-templates/service';

import { BoxSelector } from '@@/BoxSelector';
import { BoxSelectorOption } from '@@/BoxSelector/types';

import { AnalyticsStateKey } from '../types';
import { EdgeAgentTab } from '../shared/EdgeAgentTab';

import { AgentPanel } from './AgentPanel';
import { KubeConfigForm } from './KubeConfig/index';

interface Props {
  onCreate(environment: Environment, analytics: AnalyticsStateKey): void;
}

type CreationType =
  | 'agent'
  | 'edgeAgentStandard'
  | 'edgeAgentAsync'
  | 'kubeconfig';

const options: BoxSelectorOption<CreationType>[] = _.compact([
  {
    id: 'agent_endpoint',
    icon: Zap,
    iconType: 'badge',
    label: 'Agent',
    value: 'agent',
    description: '',
  },
  {
    id: 'edgeAgentStandard',
    icon: EdgeAgentStandardIcon,
    iconType: 'badge',
    label: 'Edge Agent Standard',
    description: '',
    value: 'edgeAgentStandard',
  },
  isBE && {
    id: 'edgeAgentAsync',
    icon: EdgeAgentAsyncIcon,
    iconType: 'badge',
    label: 'Edge Agent Async',
    description: '',
    value: 'edgeAgentAsync',
  },
  {
    id: 'kubeconfig_endpoint',
    icon: UploadCloud,
    iconType: 'badge',
    label: 'Import',
    value: 'kubeconfig',
    description: 'Import an existing Kubernetes config',
  },
]);

export function WizardKubernetes({ onCreate }: Props) {
  const [creationType, setCreationType] = useState(options[0].value);

  const customTemplatesQuery = useCustomTemplates();
  const customTemplates =
    customTemplatesQuery.data?.filter(
      (t) => t.Type === CustomTemplateKubernetesType
    ) || [];

  const tab = getTab(creationType, customTemplates);
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

  function getTab(type: CreationType, customTemplates: CustomTemplate[]) {
    switch (type) {
      case 'agent':
        return (
          <AgentPanel
            onCreate={(environment) => onCreate(environment, 'kubernetesAgent')}
            customTemplates={customTemplates}
          />
        );
      case 'edgeAgentStandard':
        return (
          <EdgeAgentTab
            onCreate={(environment) =>
              onCreate(environment, 'kubernetesEdgeAgentStandard')
            }
            commands={[{ ...commandsTabs.k8sLinux, label: 'Linux' }]}
          />
        );
      case 'edgeAgentAsync':
        return (
          <EdgeAgentTab
            asyncMode
            onCreate={(environment) =>
              onCreate(environment, 'kubernetesEdgeAgentAsync')
            }
            commands={[{ ...commandsTabs.k8sLinux, label: 'Linux' }]}
          />
        );
      case 'kubeconfig':
        return (
          <KubeConfigForm
            onCreate={(environment) => onCreate(environment, 'kubernetesAgent')}
            customTemplates={customTemplates}
          />
        );
      default:
        throw new Error('Creation type not supported');
    }
  }
}
