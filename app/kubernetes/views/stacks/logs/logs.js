angular.module('portainer.kubernetes').component('kubernetesStackLogsViewAngular', {
  templateUrl: './logs.html',
  controller: 'KubernetesStackLogsController',
  controllerAs: 'ctrl',
  bindings: {
    $transition$: '<',
  },
});
