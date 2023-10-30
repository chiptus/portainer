import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { IngressClassDatatableAngular } from '@/react/kubernetes/cluster/ingressClass/IngressClassDatatable/IngressClassDatatableAngular';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { NamespacesSelector } from '@/react/kubernetes/cluster/RegistryAccessView/NamespacesSelector';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { StorageAccessModeSelector } from '@/react/kubernetes/cluster/ConfigureView/ConfigureForm/StorageAccessModeSelector';
import { NamespaceAccessUsersSelector } from '@/react/kubernetes/namespaces/AccessView/NamespaceAccessUsersSelector';
import { Microk8sUpdateWidget } from '@/react/kubernetes/cluster/HomeView/Microk8sClusterDetails';
import { Microk8sClusterDetails } from '@/react/portainer/environments/ItemView/Microk8sClusterDetails';
import { NodesDatatable } from '@/react/kubernetes/cluster/HomeView/NodesDatatable';
import { RegistriesSelector } from '@/react/kubernetes/namespaces/components/RegistriesFormSection/RegistriesSelector';
import { KubeApplicationAccessPolicySelector } from '@/react/kubernetes/applications/CreateView/KubeApplicationAccessPolicySelector';
import { KubeServicesForm } from '@/react/kubernetes/applications/CreateView/application-services/KubeServicesForm';
import { kubeServicesValidation } from '@/react/kubernetes/applications/CreateView/application-services/kubeServicesValidation';
import { KubeApplicationDeploymentTypeSelector } from '@/react/kubernetes/applications/CreateView/KubeApplicationDeploymentTypeSelector';
import { Annotations } from '@/react/kubernetes/annotations';
import { YAMLReplace } from '@/react/kubernetes/components/YAMLReplace';
import { YAMLInspector } from '@/react/kubernetes/components/YAMLInspector';
import {
  ApplicationSummaryWidget,
  ApplicationDetailsWidget,
  ApplicationEventsDatatable,
} from '@/react/kubernetes/applications/DetailsView';
import { ApplicationContainersDatatable } from '@/react/kubernetes/applications/DetailsView/ApplicationContainersDatatable';
import { withFormValidation } from '@/react-tools/withFormValidation';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { ApplicationsStacksDatatable } from '@/react/kubernetes/applications/ListView/ApplicationsStacksDatatable';
import { StackName } from '@/react/kubernetes/DeployView/StackName/StackName';

export const ngModule = angular
  .module('portainer.kubernetes.react.components', [])
  .component(
    'ingressClassDatatable',
    r2a(IngressClassDatatableAngular, [
      'onChangeControllers',
      'description',
      'ingressControllers',
      'initialIngressControllers',
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
    r2a(withUIRouter(withReactQuery(withCurrentUser(RegistriesSelector))), [
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
    'kubeStackName',
    r2a(withUIRouter(withReactQuery(withCurrentUser(StackName))), [
      'setStackName',
      'isAdmin',
      'stackName',
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
    'kubeYamlInspector',
    r2a(withUIRouter(withReactQuery(withCurrentUser(YAMLInspector))), [
      'identifier',
      'data',
      'authorised',
      'system',
      'hideMessage',
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
    'applicationContainersDatatable',
    r2a(
      withUIRouter(
        withReactQuery(withCurrentUser(ApplicationContainersDatatable))
      ),
      []
    )
  )
  .component(
    'applicationDetailsWidget',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(ApplicationDetailsWidget))),
      []
    )
  )
  .component(
    'applicationEventsDatatable',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(ApplicationEventsDatatable))),
      []
    )
  )

  .component(
    'kubernetesApplicationsStacksDatatable',
    r2a(withUIRouter(withCurrentUser(ApplicationsStacksDatatable)), [
      'dataset',
      'onRefresh',
      'onRemove',
      'namespace',
      'namespaces',
      'onNamespaceChange',
      'isLoading',
      'showSystem',
      'setSystemResources',
    ])
  );

export const componentsModule = ngModule.name;

withFormValidation(
  ngModule,
  withUIRouter(withCurrentUser(withReactQuery(KubeServicesForm))),
  'kubeServicesForm',
  ['values', 'onChange', 'appName', 'selector', 'isEditMode', 'namespace'],
  kubeServicesValidation
);
