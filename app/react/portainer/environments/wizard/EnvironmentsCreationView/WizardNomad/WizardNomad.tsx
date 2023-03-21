import { useState } from 'react';
import _ from 'lodash';

import { Environment } from '@/react/portainer/environments/types';
import { commandsTabs } from '@/react/edge/components/EdgeScriptForm/scripts';
import EdgeAgentStandardIcon from '@/react/edge/components/edge-agent-standard.svg?c';
import EdgeAgentAsyncIcon from '@/react/edge/components/edge-agent-async.svg?c';

import { BoxSelector, type BoxSelectorOption } from '@@/BoxSelector';
import { BadgeIcon } from '@@/BadgeIcon';

import { AnalyticsStateKey } from '../types';
import { EdgeAgentTab } from '../shared/EdgeAgentTab';

interface Props {
  onCreate(environment: Environment, analytics: AnalyticsStateKey): void;
}

const defaultOptions: BoxSelectorOption<
  'edgeAgentStandard' | 'edgeAgentAsync'
>[] = _.compact([
  {
    id: 'edgeAgentStandard',
    icon: <BadgeIcon icon={EdgeAgentStandardIcon} size="3xl" />,
    label: 'Edge Agent Standard',
    description: '',
    value: 'edgeAgentStandard',
  },
  {
    id: 'edgeAgentAsync',
    icon: <BadgeIcon icon={EdgeAgentAsyncIcon} size="3xl" />,
    label: 'Edge Agent Async',
    description: '',
    value: 'edgeAgentAsync',
  },
]);

export function WizardNomad({ onCreate }: Props) {
  const [creationType, setCreationType] = useState(defaultOptions[0].value);

  const tab = getTab(creationType);

  return (
    <div className="form-horizontal">
      <BoxSelector
        onChange={(v) => setCreationType(v)}
        options={defaultOptions}
        value={creationType}
        radioName="creation-type"
      />
      {tab}
    </div>
  );

  function getTab(creationType: 'edgeAgentStandard' | 'edgeAgentAsync') {
    switch (creationType) {
      case 'edgeAgentStandard':
        return (
          <EdgeAgentTab
            isNomadTokenVisible
            onCreate={(environment) =>
              onCreate(environment, 'nomadEdgeAgentStandard')
            }
            commands={[commandsTabs.nomadLinux]}
          />
        );
      case 'edgeAgentAsync':
        return (
          <EdgeAgentTab
            isNomadTokenVisible
            asyncMode
            onCreate={(environment) =>
              onCreate(environment, 'nomadEdgeAgentAsync')
            }
            commands={[commandsTabs.nomadLinux]}
          />
        );
      default:
        return null;
    }
  }
}
