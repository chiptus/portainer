import { Suspense, lazy } from 'react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';

import { useCanDisplayChatbot } from '../queries';

const ChatBotItem = lazy(() =>
  import('./ChatBotItem').then(({ ChatBotItem }) => ({
    default: ChatBotItem,
  }))
);

export function LazyLoadChatbot() {
  const canDisplayChatbot = useCanDisplayChatbot();
  const environmentId = useEnvironmentId(false);

  if (!environmentId) {
    return null; // for now do not display the chat widget in views outside of environment
  }

  if (!canDisplayChatbot) {
    return null;
  }

  return (
    <Suspense fallback="">
      <ChatBotItem />
    </Suspense>
  );
}
