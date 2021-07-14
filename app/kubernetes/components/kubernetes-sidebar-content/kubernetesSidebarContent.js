angular.module('portainer.kubernetes').component('kubernetesSidebarContent', {
  templateUrl: './kubernetesSidebarContent.html',
  bindings: {
    endpointId: '<',
    adminAccess: '<',
    currentState: '<',
  },
});
