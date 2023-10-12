import { UserCheck, Link } from 'lucide-react';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';
import { useEffect } from 'react';

import { useAuthorizations } from '@/react/hooks/useUser';

import { PageHeader } from '@@/PageHeader';
import { Tab, WidgetTabs, findSelectedTabIndex } from '@@/Widget/WidgetTabs';

import { ClusterRolesDatatable } from './ClusterRolesDatatable/ClusterRolesDatatable';
import { ClusterRoleBindingsDatatable } from './ClusterRoleBindingsDatatable/ClusterRoleBindingsDatatable';

export function ClusterRolesView() {
  const router = useRouter();
  const isAuthorisedToAddOrEdit = useAuthorizations([
    'K8sClusterRoleBindingsW',
    'K8sClusterRolesW',
  ]);

  useEffect(() => {
    if (!isAuthorisedToAddOrEdit) {
      router.stateService.go('kubernetes.dashboard');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthorisedToAddOrEdit]);

  const tabs: Tab[] = [
    {
      name: 'Cluster Roles',
      icon: UserCheck,
      widget: <ClusterRolesDatatable />,
      selectedTabParam: 'clusterRoles',
    },
    {
      name: 'Cluster Role Bindings',
      icon: Link,
      widget: <ClusterRoleBindingsDatatable />,
      selectedTabParam: 'clusterRoleBindings',
    },
  ];

  const currentTabIndex = findSelectedTabIndex(
    useCurrentStateAndParams(),
    tabs
  );

  return (
    <>
      <PageHeader
        title="Cluster Role list"
        breadcrumbs="Cluster Roles"
        reload
      />
      <>
        <WidgetTabs tabs={tabs} currentTabIndex={currentTabIndex} />
        <div className="content">{tabs[currentTabIndex].widget}</div>
      </>
    </>
  );
}
