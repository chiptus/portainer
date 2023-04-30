import { useCallback } from 'react';
import { Play } from 'lucide-react';
import { useRouter } from '@uirouter/react';
import { toggleWidget } from 'react-chat-widget';

import { EnvironmentType } from '@/react/portainer/environments/types';
import { useCurrentEnvironment } from '@/react/hooks/useCurrentEnvironment';

import { Button } from '@@/buttons';

import { ChatBotResponse } from './ChatBotResponse';

const RouterStateFromEnvType: { [k in EnvironmentType]: string } = {
  [EnvironmentType.Docker]: 'docker.stacks.newstack',
  [EnvironmentType.AgentOnDocker]: 'docker.stacks.newstack',
  [EnvironmentType.EdgeAgentOnDocker]: 'docker.stacks.newstack',
  [EnvironmentType.Azure]: '',
  [EnvironmentType.KubernetesLocal]: 'kubernetes.deploy',
  [EnvironmentType.AgentOnKubernetes]: 'kubernetes.deploy',
  [EnvironmentType.EdgeAgentOnKubernetes]: 'kubernetes.deploy',
  [EnvironmentType.EdgeAgentOnNomad]: '',
};

interface Props {
  yaml: string;
}

export function ChatBotLink({ yaml }: Props) {
  const router = useRouter();

  const goTo = useCallback(
    (gotoState: string) => {
      toggleWidget();
      router.stateService.go(gotoState, { yaml });
    },
    [router, yaml]
  );

  const currentEnvQuery = useCurrentEnvironment(false);

  if (!currentEnvQuery.data) {
    return null;
  }

  const gotoState = RouterStateFromEnvType[currentEnvQuery.data.Type];
  if (!gotoState) {
    return null; // do not generate magic button for unsupported environments
  }

  return (
    <ChatBotResponse>
      <Button icon={Play} onClick={() => goTo(gotoState)}>
        Deploy in Portainer
      </Button>
    </ChatBotResponse>
  );
}
