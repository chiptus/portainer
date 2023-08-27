import { useMemo, useState } from 'react';
import YAML from 'yaml';
import { Minus, Plus } from 'lucide-react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironmentDeploymentOptions } from '@/react/portainer/environments/queries/useEnvironment';
import { useAuthorizations } from '@/react/hooks/useUser';

import { WebEditorForm } from '@@/WebEditorForm';
import { Button } from '@@/buttons';

import { YAMLReplace } from './YAMLReplace';

type Props = {
  identifier: string;
  data: string;
  authorised: boolean;
  system?: boolean;
  hideMessage?: boolean;
};

export function YAMLInspector({
  identifier,
  data,
  authorised,
  system,
  hideMessage,
}: Props) {
  const [expanded, setExpanded] = useState(false);
  const [yaml, setYaml] = useState(cleanYamlUnwantedFields(data));
  const originalYaml = useMemo(() => cleanYamlUnwantedFields(data), [data]);

  // check if the user is allowed to see edit the yaml
  const environmentId = useEnvironmentId();
  const { data: deploymentOptions } =
    useEnvironmentDeploymentOptions(environmentId);
  const roleHasAuth = useAuthorizations('K8sYAMLW');
  const isAllowedToEdit =
    roleHasAuth && !deploymentOptions?.hideWebEditor && authorised;

  return (
    <div>
      <WebEditorForm
        value={yaml}
        placeholder={
          hideMessage
            ? undefined
            : 'Define or paste the content of your manifest here'
        }
        readonly={!isAllowedToEdit || system}
        hideTitle
        onChange={(value) => setYaml(value)}
        id={identifier}
        yaml
        height={expanded ? '800px' : '500px'}
      />
      <div className="flex items-center justify-between py-5">
        <Button
          icon={expanded ? Minus : Plus}
          color="default"
          className="!ml-0"
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? 'Collapse' : 'Expand'}
        </Button>
        <YAMLReplace
          yml={yaml}
          originalYml={originalYaml}
          disabled={!isAllowedToEdit || !!system}
        />
      </div>
    </div>
  );
}

function cleanYamlUnwantedFields(yml: string) {
  try {
    const ymls = yml.split('---');
    const cleanYmls = ymls.map((yml) => {
      const y = YAML.parse(yml);
      if (y.metadata) {
        const { managedFields, resourceVersion, ...metadata } = y.metadata;
        y.metadata = metadata;
      }
      return YAML.stringify(y);
    });
    return cleanYmls.join('---\n');
  } catch (e) {
    return yml;
  }
}
