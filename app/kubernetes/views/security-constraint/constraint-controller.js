import angular from 'angular';
import './constraint.css';

angular.module('portainer.kubernetes').controller('KubernetesSecurityConstraintController', [
  '$scope',
  'EndpointProvider',
  'EndpointService',
  'OpaService',
  'Notifications',
  function ($scope, EndpointProvider, EndpointService, OpaService, Notifications) {
    $scope.state = {
      viewReady: false,
      actionInProgress: false,
    };

    $scope.formValues = {
      enabled: false,
      privilegedContainers: {
        enabled: false,
      },
      hostNamespaces: {
        enabled: false,
      },
      hostPorts: {
        enabled: false,
        hostNetwork: false,
        min: 0,
        max: 0,
      },
      volumeTypes: {
        enabled: false,
        allowedTypes: [],
      },
      hostFilesystem: {
        enabled: false,
        allowedPaths: [],
      },
      allowFlexVolumes: {
        enabled: false,
        allowedVolumes: [],
      },
      users: {
        enabled: false,
        runAsUser: {
          type: 'MustRunAs',
          idrange: [{ max: 0, min: 0 }],
        },
        runAsGroup: {
          type: 'MustRunAs',
          idrange: [{ max: 0, min: 0 }],
        },
        supplementalGroups: {
          type: 'MustRunAs',
          idrange: [{ max: 0, min: 0 }],
        },
        fsGroups: {
          type: 'MustRunAs',
          gids: [],
          idrange: [{ max: 0, min: 0 }],
        },
      },
      allowPrivilegeEscalation: {
        enabled: false,
      },
      capabilities: {
        enabled: false,
        allowedCapabilities: [],
        requiredDropCapabilities: [],
      },
      selinux: {
        enabled: false,
        // allowedCapabilities structure: { level: '', role: '', type: '', user: '' }
        allowedCapabilities: [],
      },
      allowProcMount: {
        enabled: false,
        procMountType: 'Default',
      },
      appArmor: {
        enabled: false,
        AppArmorType: ['runtime/default'],
      },
      secComp: {
        enabled: false,
        secCompType: ['runtime/default'],
      },
      forbiddenSysctlsList: {
        enabled: false,
        requiredDropCapabilities: [],
      },
    };

    $scope.endpoint = this.endpoint;
    $scope.duplicated = false;
    $scope.updateDuplicated = function () {
      $scope.duplicated =
        ($scope.formValues.volumeTypes.enabled && $scope.formValues.volumeTypes.allowedTypes.duplicated) ||
        ($scope.formValues.hostFilesystem.enabled && $scope.formValues.hostFilesystem.allowedPaths.duplicated) ||
        ($scope.formValues.allowFlexVolumes.enabled && $scope.formValues.allowFlexVolumes.allowedVolumes.duplicated) ||
        ($scope.formValues.capabilities.enabled &&
          ($scope.formValues.capabilities.duplicated ||
            $scope.formValues.capabilities.allowedCapabilities.duplicated ||
            $scope.formValues.capabilities.requiredDropCapabilities.duplicated)) ||
        ($scope.formValues.selinux.enabled && $scope.formValues.selinux.allowedCapabilities.duplicated) ||
        ($scope.formValues.appArmor.enabled && $scope.formValues.appArmor.AppArmorType.duplicated) ||
        ($scope.formValues.secComp.enabled && $scope.formValues.secComp.secCompType.duplicated) ||
        ($scope.formValues.forbiddenSysctlsList.enabled && $scope.formValues.forbiddenSysctlsList.requiredDropCapabilities.duplicated);
    };

    $scope.addItem = function (list, item) {
      if (!list) {
        list = [];
      }
      console.log(item);
      list.push(item);
    };

    $scope.removeItem = function (list, index) {
      if (list && list.length > 0) {
        list.splice(index, 1);
      }
      checkDuplicate(list);
    };

    $scope.updateItem = function (list, index, value) {
      if (list && list.length > index) {
        list[index] = value;
      }
      checkDuplicate(list);
    };

    function checkDuplicate(list) {
      list.duplicated = anyDuplicated(list);
      $scope.updateDuplicated();
    }

    function anyDuplicated(arr) {
      const validList = arr.filter((item) => {
        return item !== '';
      });
      return new Set(validList).size !== validList.length;
    }

    $scope.addHostAllowedPath = function () {
      $scope.formValues.hostFilesystem.allowedPaths.push({ pathPrefix: '', readonly: false });
      $scope.checkHostAllowedPathDuplicate();
    };

    $scope.removeHostAllowedPath = function (index) {
      const list = $scope.formValues.hostFilesystem.allowedPaths;
      if (list && list.length > 0) {
        list.splice(index, 1);
      }
      $scope.checkHostAllowedPathDuplicate();
    };

    $scope.checkHostAllowedPathDuplicate = function () {
      const list = $scope.formValues.hostFilesystem.allowedPaths.map((item) => {
        return item.pathPrefix;
      });
      $scope.formValues.hostFilesystem.allowedPaths.duplicated = anyDuplicated(list);
      $scope.updateDuplicated();
    };

    $scope.switchPrivilege = function (role) {
      if (role.idrange.length > 0) {
        return;
      }
      if (role.type === 'MustRunAs') {
        role.idrange.push({ max: 0, min: 0 });
      } else if (role.type === 'MayRunAs' && (role === $scope.formValues.users.supplementalGroups || role === $scope.formValues.users.fsGroups)) {
        role.idrange.push({ max: 0, min: 0 });
      }
    };

    $scope.removeCapability = function (list, index) {
      $scope.removeItem(list, index);
      checkCapabilityDuplicate();
    };

    $scope.updateCapability = function (list, index, value) {
      $scope.updateItem(list, index, value);
      checkCapabilityDuplicate();
    };

    function checkCapabilityDuplicate() {
      const allowedCapabilities = $scope.formValues.capabilities.allowedCapabilities.filter((item) => {
        return item !== '';
      });
      const requiredDropCapabilities = $scope.formValues.capabilities.requiredDropCapabilities.filter((item) => {
        return item !== '';
      });
      for (var i = 0; i < allowedCapabilities.length; i++) {
        for (var j = 0; j < requiredDropCapabilities.length; j++) {
          if (allowedCapabilities[i] === requiredDropCapabilities[j]) {
            $scope.formValues.capabilities.duplicated = true;
            $scope.updateDuplicated();
            return;
          }
        }
      }
      $scope.formValues.capabilities.duplicated = false;
      $scope.updateDuplicated();
    }

    $scope.removeSELinuxAllowedCapabilities = function (index) {
      const list = $scope.formValues.selinux.allowedCapabilities;
      if (list && list.length > 0) {
        list.splice(index, 1);
      }
      $scope.checkSELinuxAllowedCapabilitiesDuplicate();
    };

    $scope.checkSELinuxAllowedCapabilitiesDuplicate = function () {
      const list = $scope.formValues.selinux.allowedCapabilities.map((item) => {
        return item.level + item.role + item.type + item.user;
      });
      $scope.formValues.selinux.allowedCapabilities.duplicated = anyDuplicated(list);
      $scope.updateDuplicated();
    };

    $scope.save = function () {
      $scope.state.actionInProgress = true;
      sanitizedForm();
      OpaService.save($scope.formValues)
        .then(() => {
          Notifications.success('Constraint settings successfully saved');
          $scope.state.actionInProgress = false;
        })
        .catch((err) => {
          $scope.state.actionInProgress = false;
          Notifications.error('Failure', err, 'Unable to save constraint settings');
        });
    };

    function sanitizedForm() {
      const form = $scope.formValues;
      // Reset with default value if option is disabled.
      if (!form.hostPorts.enabled) {
        form.hostPorts.hostNetwork = false;
        form.hostPorts.min = 0;
        form.hostPorts.max = 0;
      }
      if (!form.volumeTypes.enabled) {
        form.volumeTypes.allowedTypes = [];
      }
      if (!form.hostFilesystem.enabled) {
        form.hostFilesystem.allowedPaths = [];
      }
      if (!form.allowFlexVolumes.enabled) {
        form.allowFlexVolumes.allowedVolumes = [];
      }
      if (!form.users.enabled) {
        form.users.runAsUser.idrange = [{ max: 0, min: 0 }];
        form.users.runAsGroup.idrange = [{ max: 0, min: 0 }];
        form.users.supplementalGroups.idrange = [{ max: 0, min: 0 }];
        form.users.fsGroups.idrange = [{ max: 0, min: 0 }];
      }
      if (!form.capabilities.enabled) {
        form.capabilities.allowedCapabilities = [];
        form.capabilities.requiredDropCapabilities = [];
      }
      if (!form.selinux.enabled) {
        form.selinux.allowedCapabilities = [];
      }
      if (!form.appArmor.enabled) {
        form.appArmor.AppArmorType = ['runtime/default'];
      }
      if (!form.secComp.enabled) {
        form.secComp.secCompType = ['runtime/default'];
      }
      if (!form.forbiddenSysctlsList.enabled) {
        form.forbiddenSysctlsList.requiredDropCapabilities = [];
      }
    }

    async function initView() {
      let opaData;
      const endpointID = EndpointProvider.endpointID();
      [$scope.endpoint, opaData] = await Promise.all([EndpointService.endpoint(endpointID), OpaService.detail()]);
      if (opaData.data.enabled) {
        Object.assign($scope.formValues, opaData.data);
        sanitizedForm();
      }
      $scope.state.viewReady = true;
    }

    initView();
  },
]);
