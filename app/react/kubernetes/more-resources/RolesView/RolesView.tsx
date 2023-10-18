import { useCurrentStateAndParams, useRouter } from '@uirouter/react';
import { UserCheck, Link } from 'lucide-react';
import { useEffect } from 'react';

import { useAuthorizations } from '@/react/hooks/useUser';

import { PageHeader } from '@@/PageHeader';
import { WidgetTabs, Tab, findSelectedTabIndex } from '@@/Widget/WidgetTabs';

import { RolesDatatable } from './RolesDatatable';
import { RoleBindingsDatatable } from './RoleBindingsDatatable';

export function RolesView() {
  const router = useRouter();
  const isAuthorisedToAddOrEdit = useAuthorizations([
    'K8sRoleBindingsW',
    'K8sRolesW',
  ]);

  useEffect(() => {
    if (!isAuthorisedToAddOrEdit) {
      router.stateService.go('kubernetes.dashboard');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthorisedToAddOrEdit]);

  const tabs: Tab[] = [
    {
      name: 'Roles',
      icon: UserCheck,
      widget: <RolesDatatable />,
      selectedTabParam: 'roles',
    },
    {
      name: 'Role Bindings',
      icon: Link,
      widget: <RoleBindingsDatatable />,
      selectedTabParam: 'roleBindings',
    },
  ];

  const currentTabIndex = findSelectedTabIndex(
    useCurrentStateAndParams(),
    tabs
  );

  return (
    <>
      <PageHeader title="Role list" breadcrumbs="Roles" reload />
      <>
        <WidgetTabs tabs={tabs} currentTabIndex={currentTabIndex} />
        <div className="content">{tabs[currentTabIndex].widget}</div>
      </>
    </>
  );
}
