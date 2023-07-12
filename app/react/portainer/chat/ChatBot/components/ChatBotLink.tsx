import { useCallback } from 'react';
import { Play } from 'lucide-react';
import { useRouter } from '@uirouter/react';
import { toggleWidget } from 'react-chat-widget';

import { PlatformType } from '@/react/portainer/environments/types';
import { useCurrentEnvironment } from '@/react/hooks/useCurrentEnvironment';
import { getPlatformType } from '@/react/portainer/environments/utils';
import { useAnalytics } from '@/react/hooks/useAnalytics';

import { Button } from '@@/buttons';

import { getPlatformTypeForAnalytics } from '../utils';

import { ChatBotResponse } from './ChatBotResponse';

const routerStateFromEnvType: { [k in PlatformType]: string } = {
  [PlatformType.Docker]: 'docker.stacks.newstack',
  [PlatformType.Kubernetes]: 'kubernetes.deploy',
  [PlatformType.Nomad]: '',
  [PlatformType.Azure]: '',
};

interface Props {
  yaml: string;
}

export function ChatBotLink({ yaml }: Props) {
  const router = useRouter();
  const { data: env, ...currentEnvQuery } = useCurrentEnvironment(false);
  const { trackEvent } = useAnalytics();

  const goTo = useCallback(
    (gotoState: string) => {
      trackEvent('chatbot-deploy', {
        category: 'portainer',
        metadata: {
          environmentType: getPlatformTypeForAnalytics(env?.Type),
        },
      });
      toggleWidget();
      router.stateService.go(gotoState, { yaml });
    },
    [env?.Type, router.stateService, trackEvent, yaml]
  );

  if (!env || currentEnvQuery.isLoading) {
    return null;
  }

  const gotoState = routerStateFromEnvType[getPlatformType(env.Type)];
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
