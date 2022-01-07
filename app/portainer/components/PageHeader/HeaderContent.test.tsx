import { UserContext } from '@/portainer/hooks/useUser';
import { UserViewModel } from '@/portainer/models/user';
import { renderWithQueryClient } from '@/react-tools/test-utils';

import { HeaderContainer } from './HeaderContainer';
import { HeaderContent } from './HeaderContent';

test('should not render without a wrapping HeaderContainer', async () => {
  function renderComponent() {
    return renderWithQueryClient(<HeaderContent />);
  }

  expect(renderComponent).toThrowErrorMatchingSnapshot();
});

test('should display a HeaderContent', async () => {
  const username = 'username';
  const user = new UserViewModel({ Username: username });
  const userProviderState = { user };
  const content = 'content';

  const { queryByText } = renderWithQueryClient(
    <UserContext.Provider value={userProviderState}>
      <HeaderContainer>
        <HeaderContent>{content}</HeaderContent>
      </HeaderContainer>
    </UserContext.Provider>
  );

  const contentElement = queryByText(content);
  expect(contentElement).toBeVisible();

  expect(queryByText('my account')).toBeVisible();
  expect(queryByText('log out')).toBeVisible();
});
