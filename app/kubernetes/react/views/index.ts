import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { IngressesDatatableView } from '@/react/kubernetes/ingresses/IngressDatatable';
import { CreateIngressView } from '@/react/kubernetes/ingresses/CreateIngressView';
import { LogView as ApplicationLogView } from '@/react/kubernetes/applications/LogsView';
import { LogView as StackLogView } from '@/react/kubernetes/stacks/LogsView';
import { DashboardView } from '@/react/kubernetes/DashboardView';
import { ServicesView } from '@/react/kubernetes/ServicesView';

export const viewsModule = angular
  .module('portainer.kubernetes.react.views', [])
  .component(
    'kubernetesServicesView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ServicesView))), [])
  )
  .component(
    'kubernetesIngressesView',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(IngressesDatatableView))),
      []
    )
  )
  .component(
    'kubernetesApplicationLogsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ApplicationLogView))), [])
  )
  .component(
    'kubernetesStackLogsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(StackLogView))), [
      'getLogsFn',
    ])
  )
  .component(
    'kubernetesIngressesCreateView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(CreateIngressView))), [])
  )
  .component(
    'kubernetesDashboardView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(DashboardView))), [])
  ).name;
