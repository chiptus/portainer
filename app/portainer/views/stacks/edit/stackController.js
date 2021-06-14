angular.module('portainer.app').controller('StackController', [
  '$async',
  '$q',
  '$scope',
  '$state',
  '$transition$',
  'StackService',
  'NodeService',
  'ServiceService',
  'TaskService',
  'ContainerService',
  'ServiceHelper',
  'TaskHelper',
  'Notifications',
  'FormHelper',
  'EndpointProvider',
  'EndpointService',
  'GroupService',
  'ModalService',
  'StackHelper',
  'endpoint',
  'Authentication',
  'ContainerHelper',
  function (
    $async,
    $q,
    $scope,
    $state,
    $transition$,
    StackService,
    NodeService,
    ServiceService,
    TaskService,
    ContainerService,
    ServiceHelper,
    TaskHelper,
    Notifications,
    FormHelper,
    EndpointProvider,
    EndpointService,
    GroupService,
    ModalService,
    StackHelper,
    endpoint,
    Authentication,
    ContainerHelper
  ) {
    $scope.state = {
      actionInProgress: false,
      migrationInProgress: false,
      externalStack: false,
      showEditorTab: false,
      yamlError: false,
    };

    $scope.formValues = {
      Prune: false,
      Endpoint: null,
      Env: [],
    };

    $scope.handleEnvVarChange = handleEnvVarChange;
    function handleEnvVarChange(value) {
      $scope.formValues.Env = value;
    }

    $scope.duplicateStack = function duplicateStack(name, endpointId) {
      var stack = $scope.stack;
      var env = FormHelper.removeInvalidEnvVars($scope.formValues.Env);
      EndpointProvider.setEndpointID(endpointId);

      return StackService.duplicateStack(name, $scope.stackFileContent, env, endpointId, stack.Type).then(onDuplicationSuccess).catch(notifyOnError);

      function onDuplicationSuccess() {
        Notifications.success('Stack successfully duplicated');
        $state.go('docker.stacks', {}, { reload: true });
        EndpointProvider.setEndpointID(stack.EndpointId);
      }

      function notifyOnError(err) {
        Notifications.error('Failure', err, 'Unable to duplicate stack');
      }
    };

    $scope.showEditor = function () {
      $scope.state.showEditorTab = true;
    };

    $scope.migrateStack = function (name, endpointId) {
      return $q(function (resolve) {
        ModalService.confirm({
          title: 'Are you sure?',
          message:
            'This action will deploy a new instance of this stack on the target endpoint, please note that this does NOT relocate the content of any persistent volumes that may be attached to this stack.',
          buttons: {
            confirm: {
              label: 'Migrate',
              className: 'btn-danger',
            },
          },
          callback: function onConfirm(confirmed) {
            if (!confirmed) {
              return resolve();
            }
            return resolve(migrateStack(name, endpointId));
          },
        });
      });
    };

    $scope.removeStack = function () {
      ModalService.confirmDeletion('Do you want to remove the stack? Associated services will be removed as well.', function onConfirm(confirmed) {
        if (!confirmed) {
          return;
        }
        deleteStack();
      });
    };

    function migrateStack(name, endpointId) {
      var stack = $scope.stack;
      var targetEndpointId = endpointId;

      var migrateRequest = StackService.migrateSwarmStack;
      if (stack.Type === 2) {
        migrateRequest = StackService.migrateComposeStack;
      }

      // TODO: this is a work-around for stacks created with Portainer version >= 1.17.1
      // The EndpointID property is not available for these stacks, we can pass
      // the current endpoint identifier as a part of the migrate request. It will be used if
      // the EndpointID property is not defined on the stack.
      var originalEndpointId = EndpointProvider.endpointID();
      if (stack.EndpointId === 0) {
        stack.EndpointId = originalEndpointId;
      }

      $scope.state.migrationInProgress = true;
      return migrateRequest(stack, targetEndpointId, name)
        .then(function success() {
          Notifications.success('Stack successfully migrated', stack.Name);
          $state.go('docker.stacks', {}, { reload: true });
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to migrate stack');
        })
        .finally(function final() {
          $scope.state.migrationInProgress = false;
        });
    }

    function deleteStack() {
      var endpointId = +$state.params.endpointId;
      var stack = $scope.stack;

      StackService.remove(stack, $transition$.params().external, endpointId)
        .then(function success() {
          Notifications.success('Stack successfully removed', stack.Name);
          $state.go('docker.stacks');
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to remove stack ' + stack.Name);
        });
    }

    $scope.deployStack = function () {
      var stackFile = $scope.stackFileContent;
      var env = FormHelper.removeInvalidEnvVars($scope.formValues.Env);
      var prune = $scope.formValues.Prune;
      var stack = $scope.stack;

      // TODO: this is a work-around for stacks created with Portainer version >= 1.17.1
      // The EndpointID property is not available for these stacks, we can pass
      // the current endpoint identifier as a part of the update request. It will be used if
      // the EndpointID property is not defined on the stack.
      var endpointId = EndpointProvider.endpointID();
      if (stack.EndpointId === 0) {
        stack.EndpointId = endpointId;
      }

      $scope.state.actionInProgress = true;
      StackService.updateStack(stack, stackFile, env, prune)
        .then(function success() {
          Notifications.success('Stack successfully deployed');
          $state.reload();
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to create stack');
        })
        .finally(function final() {
          $scope.state.actionInProgress = false;
        });
    };

    $scope.editorUpdate = function (cm) {
      $scope.stackFileContent = cm.getValue();
      $scope.state.yamlError = StackHelper.validateYAML($scope.stackFileContent, $scope.containerNames);
    };

    $scope.stopStack = stopStack;
    function stopStack() {
      return $async(stopStackAsync);
    }
    async function stopStackAsync() {
      const confirmed = await ModalService.confirmAsync({
        title: 'Are you sure?',
        message: 'Are you sure you want to stop this stack?',
        buttons: { confirm: { label: 'Stop', className: 'btn-danger' } },
      });
      if (!confirmed) {
        return;
      }

      $scope.state.actionInProgress = true;
      try {
        await StackService.stop($scope.stack.Id);
        $state.reload();
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to stop stack');
      }
      $scope.state.actionInProgress = false;
    }

    $scope.startStack = startStack;
    function startStack() {
      return $async(startStackAsync);
    }
    async function startStackAsync() {
      $scope.state.actionInProgress = true;
      const id = $scope.stack.Id;
      try {
        await StackService.start(id);
        $state.reload();
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to start stack');
      }
      $scope.state.actionInProgress = false;
    }

    function loadStack(id) {
      var agentProxy = $scope.applicationState.endpoint.mode.agentProxy;

      EndpointService.endpoints()
        .then(function success(data) {
          $scope.endpoints = data.value;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve endpoints');
        });

      $q.all({
        stack: StackService.stack(id),
        groups: GroupService.groups(),
        containers: ContainerService.containers(),
      })
        .then(function success(data) {
          var stack = data.stack;
          $scope.groups = data.groups;
          $scope.stack = stack;
          $scope.containerNames = ContainerHelper.getContainerNames(data.containers);

          $scope.formValues.Env = $scope.stack.Env;

          let resourcesPromise = Promise.resolve({});
          if (stack.Status === 1) {
            resourcesPromise = stack.Type === 1 ? retrieveSwarmStackResources(stack.Name, agentProxy) : retrieveComposeStackResources(stack.Name);
          }

          return $q.all({
            stackFile: StackService.getStackFile(id),
            resources: resourcesPromise,
          });
        })
        .then(function success(data) {
          $scope.stackFileContent = data.stackFile;
          if ($scope.stack.Status === 1) {
            if ($scope.stack.Type === 1) {
              assignSwarmStackResources(data.resources, agentProxy);
            } else {
              assignComposeStackResources(data.resources);
            }
          }
          $scope.state.yamlError = StackHelper.validateYAML($scope.stackFileContent, $scope.containerNames);
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve stack details');
        });
    }

    function retrieveSwarmStackResources(stackName, agentProxy) {
      var stackFilter = {
        label: ['com.docker.stack.namespace=' + stackName],
      };

      return $q.all({
        services: ServiceService.services(stackFilter),
        tasks: TaskService.tasks(stackFilter),
        containers: agentProxy ? ContainerService.containers(1) : [],
        nodes: NodeService.nodes(),
      });
    }

    function assignSwarmStackResources(resources, agentProxy) {
      var services = resources.services;
      var tasks = resources.tasks;

      if (agentProxy) {
        var containers = resources.containers;
        for (var j = 0; j < tasks.length; j++) {
          var task = tasks[j];
          TaskHelper.associateContainerToTask(task, containers);
        }
      }

      for (var i = 0; i < services.length; i++) {
        var service = services[i];
        ServiceHelper.associateTasksToService(service, tasks);
      }

      $scope.nodes = resources.nodes;
      $scope.tasks = tasks;
      $scope.services = services;
    }

    function retrieveComposeStackResources(stackName) {
      var stackFilter = {
        label: ['com.docker.compose.project=' + stackName],
      };

      return $q.all({
        containers: ContainerService.containers(1, stackFilter),
      });
    }

    function assignComposeStackResources(resources) {
      $scope.containers = resources.containers;
    }

    function loadExternalStack(name) {
      var stackType = $transition$.params().type;
      if (!stackType || (stackType !== '1' && stackType !== '2')) {
        Notifications.error('Failure', null, 'Invalid type URL parameter.');
        return;
      }

      if (stackType === '1') {
        loadExternalSwarmStack(name);
      } else {
        loadExternalComposeStack(name);
      }
    }

    function loadExternalSwarmStack(name) {
      var agentProxy = $scope.applicationState.endpoint.mode.agentProxy;

      retrieveSwarmStackResources(name, agentProxy)
        .then(function success(data) {
          assignSwarmStackResources(data, agentProxy);
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve stack details');
        });
    }

    function loadExternalComposeStack(name) {
      retrieveComposeStackResources(name)
        .then(function success(data) {
          assignComposeStackResources(data);
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve stack details');
        });
    }

    async function initView() {
      var stackName = $transition$.params().name;
      $scope.stackName = stackName;
      var external = $transition$.params().external;
      $scope.currentEndpointId = EndpointProvider.endpointID();

      if (external === 'true') {
        $scope.state.externalStack = true;
        loadExternalStack(stackName);
      } else {
        var stackId = $transition$.params().id;
        loadStack(stackId);
      }

      $scope.composeSyntaxMaxVersion = endpoint.ComposeSyntaxMaxVersion;

      $scope.stackType = $transition$.params().type;

      $scope.editorReadOnly = !Authentication.hasAuthorizations(['PortainerStackUpdate']);
    }

    initView();
  },
]);
