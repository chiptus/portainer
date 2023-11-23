import { Meta, Story } from '@storybook/react';
import { useMemo, useState } from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';

import { UserContext } from '@/react/hooks/useUser';
import { UserViewModel } from '@/portainer/models/user';
import { Role } from '@/portainer/users/types';
import { isAdmin } from '@/portainer/users/user.helpers';

import { parseAccessControlFormData } from '../utils';

import { AccessControlForm } from './AccessControlForm';

const meta: Meta = {
  title: 'Components/AccessControlForm',
  component: AccessControlForm,
};

export default meta;

const testQueryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

interface Args {
  userRole: Role;
}

function Template({ userRole }: Args) {
  const defaults = parseAccessControlFormData(isAdmin({ Role: userRole }), 0);

  const [value, setValue] = useState(defaults);

  const userProviderState = useMemo(
    () => ({ user: new UserViewModel({ Role: userRole }) }),
    [userRole]
  );

  return (
    <QueryClientProvider client={testQueryClient}>
      <UserContext.Provider value={userProviderState}>
        <AccessControlForm
          values={value}
          onChange={setValue}
          errors={{}}
          environmentId={1}
        />
      </UserContext.Provider>
    </QueryClientProvider>
  );
}

export const AdminAccessControl: Story<Args> = Template.bind({});
AdminAccessControl.args = {
  userRole: Role.Admin,
};

export const NonAdminAccessControl: Story<Args> = Template.bind({});
NonAdminAccessControl.args = {
  userRole: Role.Standard,
};
