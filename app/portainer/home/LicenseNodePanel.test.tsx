import { server, rest } from '@/setup-tests/server';
import { renderWithQueryClient } from '@/react-tools/test-utils';

import { LicenseType } from '../license-management/types';

import { LicenseNodePanel } from './LicenseNodePanel';

test('when user is using more nodes then allowed he should see message', async () => {
  const allowed = 2;
  const used = 5;
  server.use(
    rest.get('/api/licenses/info', (req, res, ctx) =>
      res(ctx.json({ nodes: allowed, type: LicenseType.Subscription }))
    ),
    rest.get('/api/status/nodes', (req, res, ctx) =>
      res(ctx.json({ nodes: used }))
    )
  );

  const { findByText } = renderWithQueryClient(<LicenseNodePanel />);

  await expect(
    findByText(
      /You have exceeded the node allowance of your license and will be unable to log into your account/
    )
  ).resolves.toBeVisible();
});

test("when user is using less nodes then allowed he shouldn't see message", async () => {
  const allowed = 5;
  const used = 2;
  server.use(
    rest.get('/api/licenses/info', (req, res, ctx) =>
      res(ctx.json({ nodes: allowed, type: LicenseType.Subscription }))
    ),
    rest.get('/api/status/nodes', (req, res, ctx) =>
      res(ctx.json({ nodes: used }))
    )
  );

  const { findByText } = renderWithQueryClient(<LicenseNodePanel />);

  await expect(
    findByText(
      /You have exceeded the node allowance of your license and will be unable to log into your account/
    )
  ).rejects.toBeTruthy();
});
