import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { IngressesDatatableView } from '@/react/kubernetes/ingresses/IngressDatatable';
import { CreateIngressView } from '@/react/kubernetes/ingresses/CreateIngressView';
import { LogView as ApplicationLogView } from '@/react/kubernetes/applications/LogsView';
import { LogView as StackLogView } from '@/react/kubernetes/stacks/LogsView';
import { DashboardView } from '@/react/kubernetes/dashboard/DashboardView';
import { ServicesView } from '@/react/kubernetes/services/ServicesView';
import { ConsoleView } from '@/react/kubernetes/applications/ConsoleView';
import { ConfigmapsAndSecretsView } from '@/react/kubernetes/configs/ListView/ConfigmapsAndSecretsView';
import { NodeCreateView } from '@/react/kubernetes/cluster/NodeCreateView/NodeCreateView';
import { Microk8sNodeStatusView } from '@/react/kubernetes/cluster/NodeStatusView';
import { NodeShellView } from '@/react/kubernetes/cluster/microk8s/NodeShell';
import { ApplicationDetailsView } from '@/react/kubernetes/applications/DetailsView/ApplicationDetailsView';

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
    'kubernetesConfigMapsAndSecretsView',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(ConfigmapsAndSecretsView))),
      []
    )
  )
  .component(
    'applicationDetailsView',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(ApplicationDetailsView))),
      []
    )
  )
  .component(
    'kubernetesNodeCreateView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(NodeCreateView))), [])
  )
  .component(
    'microk8sNodeStatusView',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(Microk8sNodeStatusView))),
      []
    )
  )
  .component(
    'microk8sNodeShellView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(NodeShellView))), [])
  )
  .component(
    'kubernetesDashboardView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(DashboardView))), [])
  )
  .component(
    'kubernetesConsoleView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ConsoleView))), [])
  ).name;
