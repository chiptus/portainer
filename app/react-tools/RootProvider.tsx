import { ReactQueryDevtools } from 'react-query/devtools';
import { QueryClient, QueryClientProvider } from 'react-query';
import { UIRouterContextComponent } from '@uirouter/react-hybrid';
import { PropsWithChildren, StrictMode } from 'react';

import { UserProvider } from '@/portainer/hooks/useUser';

const queryClient = new QueryClient();

export function RootProvider({ children }: PropsWithChildren<unknown>) {
  return (
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <UIRouterContextComponent>
          <UserProvider>{children}</UserProvider>
        </UIRouterContextComponent>
        <ReactQueryDevtools />
      </QueryClientProvider>
    </StrictMode>
  );
}
