import { fireEvent } from '@testing-library/react';

import { makeTestRouter } from '@/react/test-utils/makeTestRouter';

import { AddButton } from './AddButton';

// const Component = withTestUIRouter(AddButton, );

const states = [
  { name: 'root', url: '/root' },
  { name: 'new', url: '/new', parent: 'root' },
];

let { routerGo, mountInRouter } = makeTestRouter([]);
beforeEach(() => {
  ({ routerGo, mountInRouter } = makeTestRouter(states));
});
function renderDefault({ label = 'default label' }: { label?: string } = {}) {
  return mountInRouter(<AddButton>{label}</AddButton>);
}

test('should display a AddButton component and allow onClick', async () => {
  const label = 'test label';
  await routerGo('root');
  const { findByText } = renderDefault({ label });

  const buttonLabel = await findByText(label);
  expect(buttonLabel).toBeTruthy();

  fireEvent.click(buttonLabel);
});
