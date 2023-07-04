import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { IngressClassDatatable } from '@/react/kubernetes/cluster/ingressClass/IngressClassDatatable';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { NamespacesSelector } from '@/react/kubernetes/cluster/RegistryAccessView/NamespacesSelector';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { StorageAccessModeSelector } from '@/react/kubernetes/cluster/ConfigureView/StorageAccessModeSelector';
import { NamespaceAccessUsersSelector } from '@/react/kubernetes/namespaces/AccessView/NamespaceAccessUsersSelector';
import { Microk8sUpdateWidget } from '@/react/kubernetes/cluster/HomeView/Microk8sClusterDetails';
import { Microk8sClusterDetails } from '@/react/portainer/environments/ItemView/Microk8sClusterDetails';
import { NodesDatatable } from '@/react/kubernetes/cluster/HomeView/NodesDatatable';
import { CreateNamespaceRegistriesSelector } from '@/react/kubernetes/namespaces/CreateView/CreateNamespaceRegistriesSelector';
import { KubeApplicationAccessPolicySelector } from '@/react/kubernetes/applications/CreateView/KubeApplicationAccessPolicySelector';
import { KubeServicesForm } from '@/react/kubernetes/applications/CreateView/application-services/KubeServicesForm';
import { kubeServicesValidation } from '@/react/kubernetes/applications/CreateView/application-services/kubeServicesValidation';
import { KubeApplicationDeploymentTypeSelector } from '@/react/kubernetes/applications/CreateView/KubeApplicationDeploymentTypeSelector';
import { Annotations } from '@/react/kubernetes/annotations';
import { YAMLReplace } from '@/react/kubernetes/common/YAMLReplace';
import {
  ApplicationSummaryWidget,
  ApplicationDetailsWidget,
} from '@/react/kubernetes/applications/DetailsView';
import { withUserProvider } from '@/react/test-utils/withUserProvider';
import { withFormValidation } from '@/react-tools/withFormValidation';
import { withCurrentUser } from '@/react-tools/withCurrentUser';

export const ngModule = angular
  .module('portainer.kubernetes.react.components', [])
  .component(
    'ingressClassDatatable',
    r2a(IngressClassDatatable, [
      'onChangeControllers',
      'description',
      'ingressControllers',
      'allowNoneIngressClass',
      'isLoading',
      'noIngressControllerLabel',
      'view',
    ])
  )
  .component(
    'namespacesSelector',
    r2a(NamespacesSelector, [
      'dataCy',
      'inputId',
      'name',
      'namespaces',
      'onChange',
      'placeholder',
      'value',
    ])
  )
  .component(
    'storageAccessModeSelector',
    r2a(StorageAccessModeSelector, [
      'inputId',
      'onChange',
      'options',
      'value',
      'storageClassName',
    ])
  )
  .component(
    'namespaceAccessUsersSelector',
    r2a(NamespaceAccessUsersSelector, [
      'inputId',
      'onChange',
      'options',
      'value',
      'dataCy',
      'placeholder',
      'name',
    ])
  )
  .component(
    'microk8sUpdateWidget',
    r2a(withUIRouter(withReactQuery(withCurrentUser(Microk8sUpdateWidget))), [])
  )
  .component(
    'microk8sClusterDetails',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(Microk8sClusterDetails))),
      []
    )
  )
  .component(
    'kubeNodesDatatable',
    r2a(withUIRouter(withReactQuery(withCurrentUser(NodesDatatable))), [])
  )
  .component(
    'createNamespaceRegistriesSelector',
    r2a(CreateNamespaceRegistriesSelector, [
      'inputId',
      'onChange',
      'options',
      'value',
    ])
  )
  .component(
    'kubeApplicationAccessPolicySelector',
    r2a(KubeApplicationAccessPolicySelector, [
      'value',
      'onChange',
      'isEdit',
      'persistedFoldersUseExistingVolumes',
    ])
  )
  .component(
    'kubeApplicationDeploymentTypeSelector',
    r2a(KubeApplicationDeploymentTypeSelector, [
      'value',
      'onChange',
      'supportGlobalDeployment',
    ])
  )
  .component(
    'annotations',
    r2a(Annotations, [
      'initialAnnotations',
      'hideForm',
      'errors',
      'placeholder',
      'ingressType',
      'handleUpdateAnnotations',
      'screen',
      'index',
    ])
  )
  .component(
    'yamlReplace',
    r2a(withUIRouter(withReactQuery(withCurrentUser(YAMLReplace))), [
      'yml',
      'originalYml',
      'disabled',
    ])
  )
  .component(
    'applicationSummaryWidget',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(ApplicationSummaryWidget))),
      []
    )
  )
  .component(
    'applicationDetailsWidget',
    r2a(
      withUIRouter(withReactQuery(withUserProvider(ApplicationDetailsWidget))),
      []
    )
  );

export const componentsModule = ngModule.name;

withFormValidation(
  ngModule,
  withUIRouter(withCurrentUser(withReactQuery(KubeServicesForm))),
  'kubeServicesForm',
  ['values', 'onChange', 'appName', 'selector', 'isEditMode', 'namespace'],
  kubeServicesValidation
);
