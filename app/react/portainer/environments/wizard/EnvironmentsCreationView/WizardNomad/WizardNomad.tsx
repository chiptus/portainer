import { useState } from 'react';
import { Cloud } from 'lucide-react';

import {
  Environment,
  EnvironmentCreationTypes,
} from '@/react/portainer/environments/types';
import {
  buildLinuxNomadCommand,
  CommandTab,
} from '@/react/edge/components/EdgeScriptForm/scripts';

import { BoxSelector } from '@@/BoxSelector';
import { BadgeIcon } from '@@/BadgeIcon';

import { AnalyticsStateKey } from '../types';
import { EdgeAgentTab } from '../shared/EdgeAgentTab';
import { useFilterEdgeOptionsIfNeeded } from '../useOnlyEdgeOptions';

interface Props {
  onCreate(environment: Environment, analytics: AnalyticsStateKey): void;
}

const commands: CommandTab[] = [
  {
    id: 'nomad',
    label: 'Linux',
    command: buildLinuxNomadCommand,
  },
];

const defaultOptions = [
  {
    description: 'Portainer Edge Agent',
    icon: <BadgeIcon icon={Cloud} size="3xl" />,
    id: 'id',
    label: 'Edge Agent',
    value: EnvironmentCreationTypes.EdgeAgentEnvironment,
  },
];

export function WizardNomad({ onCreate }: Props) {
  const options = useFilterEdgeOptionsIfNeeded(
    defaultOptions,
    EnvironmentCreationTypes.EdgeAgentEnvironment
  );

  const [selected, setSelected] = useState(options[0].value);

  return (
    <div className="form-horizontal">
      <BoxSelector
        onChange={(value) => setSelected(value)}
        options={options}
        value={selected}
        radioName="creation-type"
      />

      <EdgeAgentTab
        commands={commands}
        onCreate={(environment) => onCreate(environment, 'nomadEdgeAgent')}
        isNomadTokenVisible
      />
    </div>
  );
}
