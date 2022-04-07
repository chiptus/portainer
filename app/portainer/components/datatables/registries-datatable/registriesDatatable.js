angular.module('portainer.app').component('registriesDatatable', {
  templateUrl: './registriesDatatable.html',
  controller: 'RegistriesDatatableController',
  bindings: {
    titleText: '@',
    titleIcon: '@',
    dataset: '<',
    endpointId: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    removeAction: '<',
    canBrowse: '<',
    endpointType: '<',
    canManageAccess: '<',
  },
});
