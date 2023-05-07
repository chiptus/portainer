import { Environment } from '@/react/portainer/environments/types';
import { CustomTemplate } from '@/react/portainer/custom-templates/types';

import { AgentForm } from '../shared/AgentForm';

import { DeploymentScripts } from './DeploymentScripts';

interface Props {
  onCreate(environment: Environment): void;
  customTemplates: CustomTemplate[];
}

export function AgentPanel({ onCreate, customTemplates }: Props) {
  return (
    <>
      <DeploymentScripts />

      <div className="mt-5">
        <AgentForm onCreate={onCreate} customTemplates={customTemplates} />
      </div>
    </>
  );
}
