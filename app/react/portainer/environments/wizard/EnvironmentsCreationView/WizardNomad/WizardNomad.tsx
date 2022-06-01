import { Environment } from '@/portainer/environments/types';
import { BoxSelector } from '@/portainer/components/BoxSelector';
import {
  buildLinuxNomadCommand,
  CommandTab,
} from '@/react/edge/components/EdgeScriptForm/scripts';

import { AnalyticsStateKey } from '../types';
import { EdgeAgentTab } from '../shared/EdgeAgentTab';

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

const options = [
  {
    description: 'Portainer Edge Agent',
    icon: 'fa fa-cloud',
    id: 'id',
    label: 'Edge Agent',
    value: 'edge',
  },
];

export function WizardNomad({ onCreate }: Props) {
  return (
    <div className="form-horizontal">
      <div className="form-group">
        <div className="col-sm-12">
          <BoxSelector
            onChange={() => {}}
            options={options}
            value="edge"
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
