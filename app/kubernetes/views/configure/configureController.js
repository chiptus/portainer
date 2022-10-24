import _ from 'lodash-es';
import angular from 'angular';
import { KubernetesStorageClass, KubernetesStorageClassAccessPolicies } from 'Kubernetes/models/storage-class/models';
import { KubernetesFormValidationReferences } from 'Kubernetes/models/application/formValues';
import { KubernetesIngressClassTypes } from 'Kubernetes/ingress/constants';
import KubernetesNamespaceHelper from 'Kubernetes/helpers/namespaceHelper';
import { FeatureId } from '@/portainer/feature-flags/enums';

import { getIngressControllerClassMap, updateIngressControllerClassMap } from '@/react/kubernetes/cluster/ingressClass/utils';

class KubernetesConfigureController {
  /* #region  CONSTRUCTOR */

  /* @ngInject */
  constructor(
    $analytics,
    $async,
    $state,
    $scope,
    Notifications,
    KubernetesStorageService,
    EndpointService,
    EndpointProvider,
    ModalService,
    KubernetesResourcePoolService,
    KubernetesIngressService,
    KubernetesMetricsService
  ) {
    this.$analytics = $analytics;
    this.$async = $async;
    this.$state = $state;
    this.$scope = $scope;
    this.Notifications = Notifications;
    this.KubernetesStorageService = KubernetesStorageService;
    this.EndpointService = EndpointService;
    this.EndpointProvider = EndpointProvider;
    this.ModalService = ModalService;
    this.KubernetesResourcePoolService = KubernetesResourcePoolService;
    this.KubernetesIngressService = KubernetesIngressService;
    this.KubernetesMetricsService = KubernetesMetricsService;

    this.IngressClassTypes = KubernetesIngressClassTypes;

    this.onInit = this.onInit.bind(this);
    this.configureAsync = this.configureAsync.bind(this);
    this.areControllersChanged = this.areControllersChanged.bind(this);
    this.areFormValuesChanged = this.areFormValuesChanged.bind(this);
    this.onBeforeOnload = this.onBeforeOnload.bind(this);
    this.limitedFeature = FeatureId.K8S_SETUP_DEFAULT;
    this.limitedFeatureAutoWindow = FeatureId.HIDE_AUTO_UPDATE_WINDOW;
    this.onToggleAutoUpdate = this.onToggleAutoUpdate.bind(this);
    this.onChangeControllers = this.onChangeControllers.bind(this);
    this.onChangeEnableResourceOverCommit = this.onChangeEnableResourceOverCommit.bind(this);
    this.onToggleIngressAvailabilityPerNamespace = this.onToggleIngressAvailabilityPerNamespace.bind(this);
    this.onToggleAllowNoneIngressClass = this.onToggleAllowNoneIngressClass.bind(this);
    this.onChangeStorageClassAccessMode = this.onChangeStorageClassAccessMode.bind(this);
  }
  /* #endregion */

  /* #region  STORAGE CLASSES UI MANAGEMENT */
  storageClassAvailable() {
    return this.StorageClasses && this.StorageClasses.length > 0;
  }

  hasValidStorageConfiguration() {
    let valid = true;
    _.forEach(this.StorageClasses, (item) => {
      if (item.selected && item.AccessModes.length === 0) {
        valid = false;
      }
    });
    return valid;
  }
  /* #endregion */

  /* #region  INGRESS CLASSES UI MANAGEMENT */
  onChangeControllers(controllerClassMap) {
    this.ingressControllers = controllerClassMap;
  }

  hasTraefikIngress() {
    return _.find(this.formValues.IngressClasses, { Type: this.IngressClassTypes.TRAEFIK });
  }

  toggleAdvancedIngSettings() {
    this.$scope.$evalAsync(() => {
      this.state.isIngToggleSectionExpanded = !this.state.isIngToggleSectionExpanded;
    });
  }

  onToggleAllowNoneIngressClass() {
    this.$scope.$evalAsync(() => {
      this.formValues.AllowNoneIngressClass = !this.formValues.AllowNoneIngressClass;
    });
  }

  onToggleIngressAvailabilityPerNamespace() {
    this.$scope.$evalAsync(() => {
      this.formValues.IngressAvailabilityPerNamespace = !this.formValues.IngressAvailabilityPerNamespace;
    });
  }
  /* #endregion */

  /* #region RESOURCES AND METRICS */

  onChangeEnableResourceOverCommit(enabled) {
    this.$scope.$evalAsync(() => {
      this.formValues.EnableResourceOverCommit = enabled;
      if (enabled) {
        this.formValues.ResourceOverCommitPercentage = 20;
      }
    });
  }

  /* #endregion */

  /* #region  CONFIGURE */
  assignFormValuesToEndpoint(endpoint, storageClasses, ingressClasses) {
    endpoint.Kubernetes.Configuration.StorageClasses = storageClasses;
    endpoint.Kubernetes.Configuration.UseLoadBalancer = this.formValues.UseLoadBalancer;
    endpoint.Kubernetes.Configuration.UseServerMetrics = this.formValues.UseServerMetrics;
    endpoint.Kubernetes.Configuration.EnableResourceOverCommit = this.formValues.EnableResourceOverCommit;
    endpoint.Kubernetes.Configuration.ResourceOverCommitPercentage = this.formValues.ResourceOverCommitPercentage;
    endpoint.Kubernetes.Configuration.IngressClasses = ingressClasses;
    endpoint.Kubernetes.Configuration.RestrictDefaultNamespace = this.formValues.RestrictDefaultNamespace;
    endpoint.Kubernetes.Configuration.IngressAvailabilityPerNamespace = this.formValues.IngressAvailabilityPerNamespace;
    endpoint.Kubernetes.Configuration.AllowNoneIngressClass = this.formValues.AllowNoneIngressClass;
    endpoint.ChangeWindow = this.state.autoUpdateSettings;
  }

  transformFormValues() {
    const storageClasses = _.map(this.StorageClasses, (item) => {
      if (item.selected) {
        const res = new KubernetesStorageClass();
        res.Name = item.Name;
        res.AccessModes = _.map(item.AccessModes, 'Name');
        res.Provisioner = item.Provisioner;
        res.AllowVolumeExpansion = item.AllowVolumeExpansion;
        return res;
      }
    });
    _.pull(storageClasses, undefined);

    const ingressClasses = _.without(
      _.map(this.formValues.IngressClasses, (ic) => (ic.NeedsDeletion ? undefined : ic)),
      undefined
    );
    _.pull(ingressClasses, undefined);

    return [storageClasses, ingressClasses];
  }

  async removeIngressesAcrossNamespaces() {
    const ingressesToDel = _.filter(this.formValues.IngressClasses, { NeedsDeletion: true });
    if (!ingressesToDel.length) {
      return;
    }
    const promises = [];
    const oldEndpointID = this.EndpointProvider.endpointID();
    this.EndpointProvider.setEndpointID(this.endpoint.Id);

    try {
      const allResourcePools = await this.KubernetesResourcePoolService.get();
      const resourcePools = _.filter(
        allResourcePools,
        (resourcePool) => !KubernetesNamespaceHelper.isSystemNamespace(resourcePool.Namespace.Name) && !KubernetesNamespaceHelper.isDefaultNamespace(resourcePool.Namespace.Name)
      );

      ingressesToDel.forEach((ingress) => {
        resourcePools.forEach((resourcePool) => {
          promises.push(this.KubernetesIngressService.delete(resourcePool.Namespace.Name, ingress.Name));
        });
      });
    } finally {
      this.EndpointProvider.setEndpointID(oldEndpointID);
    }

    const responses = await Promise.allSettled(promises);
    responses.forEach((respons) => {
      if (respons.status == 'rejected' && respons.reason.err.status != 404) {
        throw respons.reason;
      }
    });
  }

  enableMetricsServer() {
    if (this.formValues.UseServerMetrics) {
      this.state.metrics.userClick = true;
      this.state.metrics.pending = true;
      this.KubernetesMetricsService.capabilities(this.endpoint.Id)
        .then(() => {
          this.state.metrics.isServerRunning = true;
          this.state.metrics.pending = false;
          this.formValues.UseServerMetrics = true;
        })
        .catch(() => {
          this.state.metrics.isServerRunning = false;
          this.state.metrics.pending = false;
          this.formValues.UseServerMetrics = false;
        });
    } else {
      this.state.metrics.userClick = false;
      this.formValues.UseServerMetrics = false;
    }
  }

  async configureAsync() {
    try {
      this.state.actionInProgress = true;
      const [storageClasses, ingressClasses] = this.transformFormValues();

      await this.removeIngressesAcrossNamespaces();

      this.assignFormValuesToEndpoint(this.endpoint, storageClasses, ingressClasses);
      await this.EndpointService.updateEndpoint(this.endpoint.Id, this.endpoint);
      // updateIngressControllerClassMap must be done after updateEndpoint, as a hacky workaround. A better solution: saving ingresscontrollers somewhere else, is being discussed
      await updateIngressControllerClassMap(this.state.endpointId, this.ingressControllers || []);
      this.state.isSaving = true;
      const storagePromises = _.map(storageClasses, (storageClass) => {
        const oldStorageClass = _.find(this.oldStorageClasses, { Name: storageClass.Name });
        if (oldStorageClass) {
          return this.KubernetesStorageService.patch(this.state.endpointId, oldStorageClass, storageClass);
        }
      });
      await Promise.all(storagePromises);

      const endpoints = this.EndpointProvider.endpoints();
      const modifiedEndpoint = _.find(endpoints, (item) => item.Id === this.endpoint.Id);
      if (modifiedEndpoint) {
        this.assignFormValuesToEndpoint(modifiedEndpoint, storageClasses, ingressClasses);
        this.EndpointProvider.setEndpoints(endpoints);
      }
      this.Notifications.success('Success', 'Configuration successfully applied');
      this.$state.go('portainer.home');
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to apply configuration');
    } finally {
      this.state.actionInProgress = false;

      // Timezone is only for Analytics, not for API payload
      if (this.state.autoUpdateSettings) {
        this.$analytics.eventTrack('time-window-create', {
          category: 'kubernetes',
          metadata: {
            'Start-time': this.state.autoUpdateSettings.StartTime,
            'End-time': this.state.autoUpdateSettings.EndTime,
            'Time-zone': this.state.timeZone,
          },
        });
      }
    }
  }

  configure() {
    return this.$async(this.configureAsync);
  }
  /* #endregion */

  restrictDefaultToggledOn() {
    return this.formValues.RestrictDefaultNamespace && !this.oldFormValues.RestrictDefaultNamespace;
  }

  onToggleAutoUpdate(value) {
    return this.$scope.$evalAsync(() => {
      this.state.autoUpdateSettings.Enabled = value;
    });
  }

  onChangeStorageClassAccessMode(storageClassName, accessModes) {
    return this.$scope.$evalAsync(() => {
      const storageClass = this.StorageClasses.find((item) => item.Name === storageClassName);

      if (!storageClass) {
        throw new Error('Storage class not found');
      }

      storageClass.AccessModes = accessModes;
    });
  }

  /* #region  ON INIT */
  async onInit() {
    this.state = {
      actionInProgress: false,
      displayConfigureClassPanel: {},
      viewReady: false,
      isIngToggleSectionExpanded: false,
      endpointId: this.$state.params.endpointId,
      duplicates: {
        ingressClasses: new KubernetesFormValidationReferences(),
      },
      metrics: {
        pending: false,
        isServerRunning: false,
        userClick: false,
      },
      timeZone: '',
      isSaving: false,
    };

    this.formValues = {
      UseLoadBalancer: false,
      UseServerMetrics: false,
      EnableResourceOverCommit: true,
      ResourceOverCommitPercentage: 20,
      IngressClasses: [],
      RestrictDefaultNamespace: false,
      enableAutoUpdateTimeWindow: false,
      IngressAvailabilityPerNamespace: false,
    };

    this.isIngressControllersLoading = true;
    try {
      [this.StorageClasses, this.endpoint] = await Promise.all([this.KubernetesStorageService.get(this.state.endpointId), this.EndpointService.endpoint(this.state.endpointId)]);

      this.ingressControllers = await getIngressControllerClassMap({ environmentId: this.state.endpointId });
      this.originalIngressControllers = structuredClone(this.ingressControllers);

      this.state.autoUpdateSettings = this.endpoint.ChangeWindow;

      this.availableAccessModes = new KubernetesStorageClassAccessPolicies();
      _.forEach(this.StorageClasses, (item) => {
        const storage = _.find(this.endpoint.Kubernetes.Configuration.StorageClasses, (sc) => sc.Name === item.Name);
        if (storage) {
          item.selected = true;
          item.AccessModes = storage.AccessModes.map((name) => this.availableAccessModes.find((accessMode) => accessMode.Name === name));
        } else if (this.availableAccessModes.length) {
          // set a default access mode if the storage class is not enabled and there are available access modes
          item.AccessModes = [this.availableAccessModes[0]];
        }
      });

      this.oldStorageClasses = angular.copy(this.StorageClasses);

      this.formValues.UseLoadBalancer = this.endpoint.Kubernetes.Configuration.UseLoadBalancer;
      this.formValues.UseServerMetrics = this.endpoint.Kubernetes.Configuration.UseServerMetrics;
      this.formValues.EnableResourceOverCommit = this.endpoint.Kubernetes.Configuration.EnableResourceOverCommit;
      this.formValues.ResourceOverCommitPercentage = this.endpoint.Kubernetes.Configuration.ResourceOverCommitPercentage;
      this.formValues.RestrictDefaultNamespace = this.endpoint.Kubernetes.Configuration.RestrictDefaultNamespace;
      this.formValues.IngressClasses = _.map(this.endpoint.Kubernetes.Configuration.IngressClasses, (ic) => {
        ic.IsNew = false;
        ic.NeedsDeletion = false;
        return ic;
      });
      this.formValues.IngressAvailabilityPerNamespace = this.endpoint.Kubernetes.Configuration.IngressAvailabilityPerNamespace;
      this.formValues.AllowNoneIngressClass = this.endpoint.Kubernetes.Configuration.AllowNoneIngressClass;

      this.oldFormValues = Object.assign({}, this.formValues);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve environment configuration');
    } finally {
      this.state.viewReady = true;
      this.isIngressControllersLoading = false;
    }

    window.addEventListener('beforeunload', this.onBeforeOnload);
  }

  $onInit() {
    return this.$async(this.onInit);
  }
  /* #endregion */

  $onDestroy() {
    window.removeEventListener('beforeunload', this.onBeforeOnload);
  }

  areControllersChanged() {
    return !_.isEqual(this.ingressControllers, this.originalIngressControllers);
  }

  areFormValuesChanged() {
    return !_.isEqual(this.formValues, this.oldFormValues);
  }

  onBeforeOnload(event) {
    if (!this.state.isSaving && (this.areControllersChanged() || this.areFormValuesChanged())) {
      event.preventDefault();
      event.returnValue = '';
    }
  }

  uiCanExit() {
    if (!this.state.isSaving && (this.areControllersChanged() || this.areFormValuesChanged()) && !this.isIngressControllersLoading) {
      return this.ModalService.confirmAsync({
        title: 'Are you sure?',
        message: 'You currently have unsaved changes in the cluster setup view. Are you sure you want to leave?',
        buttons: {
          confirm: {
            label: 'Yes',
            className: 'btn-danger',
          },
        },
      });
    }
  }
}

export default KubernetesConfigureController;
angular.module('portainer.kubernetes').controller('KubernetesConfigureController', KubernetesConfigureController);
