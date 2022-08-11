import { useState } from 'react';

import {
  Environment,
  EnvironmentCreationTypes,
} from '@/portainer/environments/types';
import {
  buildLinuxNomadCommand,
  CommandTab,
} from '@/react/edge/components/EdgeScriptForm/scripts';

import { BoxSelector } from '@@/BoxSelector';

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
    icon: 'svg-edgeagent',
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
      <div className="form-group">
        <div className="col-sm-12">
          <BoxSelector
            onChange={(value) => setSelected(value)}
            options={options}
            value={selected}
            radioName="creation-type"
          />
        </div>
      </div>

      <EdgeAgentTab
        commands={commands}
        onCreate={(environment) => onCreate(environment, 'nomadEdgeAgent')}
        isNomadTokenVisible
      />
    </div>
  );
}
