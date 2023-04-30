import { useCallback, useEffect } from 'react';
import {
  Widget,
  addResponseMessage,
  toggleMsgLoader,
  toggleInputDisabled,
  renderCustomComponent,
} from 'react-chat-widget';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import smallLogo from '@/assets/ico/logomark.svg';

import { useChatQueryMutation, useCanDisplayChatbot } from '../queries';
import { ChatQueryContext } from '../types';

import { ChatBotLink } from './components/ChatBotLink';
import 'react-chat-widget/lib/styles.css';
import styles from './ChatBot.module.css';

function toggleWaitingState() {
  toggleInputDisabled();
  toggleMsgLoader();
}

export function ChatBotItem() {
  const askMutation = useChatQueryMutation();
  const environmentId = useEnvironmentId(false);
  const canDisplayChatbot = useCanDisplayChatbot();

  const handleNewUserMessage = useCallback(
    async (newMessage: string) => {
      toggleWaitingState();
      askMutation.mutate(
        {
          Message: newMessage,
          Context: ChatQueryContext.ENVIRONMENT_AWARE,
          EnvironmentID: environmentId,
        },
        {
          onSuccess: ({ message, yaml }) => {
            addResponseMessage(message);
            if (yaml && yaml !== '') {
              renderCustomComponent(ChatBotLink, { yaml });
            }
          },
          onError: (err) => {
            const e = err as Error;
            addResponseMessage(`An error occurred: ${e.message}

You can check the status of the OpenAI API at [https://status.openai.com](https://status.openai.com)`);
          },
          onSettled: () => {
            toggleWaitingState();
          },
        }
      );
    },
    [askMutation, environmentId]
  );

  useEffect(() => {
    sendDisclaimer();
  }, []);

  if (!environmentId) {
    return null; // for now do not display the chat widget in views outside of environment
  }

  if (!canDisplayChatbot) {
    return null;
  }

  return (
    <div className={styles.root}>
      <Widget
        handleNewUserMessage={handleNewUserMessage}
        profileAvatar={smallLogo}
        title="Portainer Assistant"
        launcherOpenImg={smallLogo}
        subtitle="This is an experimental feature. Responses might be inaccurate."
        showBadge={false}
      />
    </div>
  );
}

function sendDisclaimer() {
  addResponseMessage(`This experimental feature is powered by OpenAI. Please note:
  * Portainer does not store any information related to the chat data.
  * Portainer does not forward sensitive information to OpenAI, outside of what you might query in the chat.
  * Chat responses may contain inaccurate information.
  * Response times can be slow, due to the nature of the OpenAI API.
  `);
  addResponseMessage(
    '**Warning: Do not share sensitive information in the chat, as OpenAI may retain it.**'
  );
  addResponseMessage('How can I help you?');
}
