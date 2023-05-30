import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { IngressClassDatatable } from '@/react/kubernetes/cluster/ingressClass/IngressClassDatatable';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { NamespacesSelector } from '@/react/kubernetes/cluster/RegistryAccessView/NamespacesSelector';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { StorageAccessModeSelector } from '@/react/kubernetes/cluster/ConfigureView/StorageAccessModeSelector';
import { NamespaceAccessUsersSelector } from '@/react/kubernetes/namespaces/AccessView/NamespaceAccessUsersSelector';
import { Microk8sClusterDetails } from '@/react/kubernetes/cluster/Microk8sClusterDetails';
import { CreateNamespaceRegistriesSelector } from '@/react/kubernetes/namespaces/CreateView/CreateNamespaceRegistriesSelector';
import { KubeApplicationAccessPolicySelector } from '@/react/kubernetes/applications/CreateView/KubeApplicationAccessPolicySelector';
import { KubeApplicationDeploymentTypeSelector } from '@/react/kubernetes/applications/CreateView/KubeApplicationDeploymentTypeSelector';
import { Annotations } from '@/react/kubernetes/annotations';
import { YAMLReplace } from '@/react/kubernetes/common/YAMLReplace';
import {
  ApplicationSummaryWidget,
  ApplicationDetailsWidget,
} from '@/react/kubernetes/applications/DetailsView';
import { withUserProvider } from '@/react/test-utils/withUserProvider';

export const componentsModule = angular
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
    'microk8sClusterDetails',
    r2a(withUIRouter(withReactQuery(Microk8sClusterDetails)), [])
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
    r2a(withUIRouter(withReactQuery(withUserProvider(YAMLReplace))), [
      'yml',
      'originalYml',
      'disabled',
    ])
  )
  .component(
    'applicationSummaryWidget',
    r2a(
      withUIRouter(withReactQuery(withUserProvider(ApplicationSummaryWidget))),
      []
    )
  )
  .component(
    'applicationDetailsWidget',
    r2a(
      withUIRouter(withReactQuery(withUserProvider(ApplicationDetailsWidget))),
      []
    )
  ).name;
