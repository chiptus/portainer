import { PropsWithChildren } from 'react';

export function ChatBotResponse({ children }: PropsWithChildren<unknown>) {
  return (
    <>
      <div className="rcw-avatar" />
      <div className="rcw-response">{children}</div>
    </>
  );
}
