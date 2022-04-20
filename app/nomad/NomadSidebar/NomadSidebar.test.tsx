import { render, within } from '@/react-tools/test-utils';

import { NomadSidebar } from './NomadSidebar';

test('dashboard items should render correctly', () => {
  const { getByLabelText } = renderComponent();
  const dashboardItem = getByLabelText('Dashboard');
  expect(dashboardItem).toBeVisible();
  expect(dashboardItem).toHaveTextContent('Dashboard');

  const dashboardItemElements = within(dashboardItem);
  expect(dashboardItemElements.getByLabelText('itemIcon')).toBeVisible();
  expect(dashboardItemElements.getByLabelText('itemIcon')).toHaveClass(
    'fa-tachometer-alt',
    'fa-fw'
  );

  const jobsItem = getByLabelText('Jobs');
  expect(jobsItem).toBeVisible();
  expect(jobsItem).toHaveTextContent('Jobs');

  const jobsItemElements = within(jobsItem);
  expect(jobsItemElements.getByLabelText('itemIcon')).toBeVisible();
  expect(jobsItemElements.getByLabelText('itemIcon')).toHaveClass(
    'fa-th-list',
    'fa-fw'
  );
});

function renderComponent() {
  return render(<NomadSidebar environmentId="1" />);
}
