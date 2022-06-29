import getEndpointsTotalCount from './transform/getEndpointsTotalCount';

angular.module('portainer.app').factory('Endpoints', [
  '$resource',
  'API_ENDPOINT_ENDPOINTS',
  function EndpointsFactory($resource, API_ENDPOINT_ENDPOINTS) {
    'use strict';
    return $resource(
      API_ENDPOINT_ENDPOINTS + '/:id/:action',
      {},
      {
        query: {
          method: 'GET',
          params: { start: '@start', limit: '@limit', search: '@search', groupId: '@groupId' },
          transformResponse: getEndpointsTotalCount,
        },
        get: { method: 'GET', params: { id: '@id' } },
        update: { method: 'PUT', params: { id: '@id' } },
        disassociate: { method: 'DELETE', params: { id: '@id', action: 'association' } },
        updateAccess: { method: 'PUT', params: { id: '@id', action: 'access' } },
        remove: { method: 'DELETE', params: { id: '@id' } },
        snapshots: { method: 'POST', params: { action: 'snapshot' } },
        dashboardQuery: { method: 'POST', params: { id: '@id', action: 'snapshot' } },
        status: { method: 'GET', params: { id: '@id', action: 'status' } },
        updateSettings: { method: 'PUT', params: { id: '@id', action: 'settings' } },
        dockerhubLimits: { method: 'GET', url: `${API_ENDPOINT_ENDPOINTS}/:id/dockerhub/:registryId` },
        updatePoolAccess: { url: API_ENDPOINT_ENDPOINTS + '/:id/pools/:rpn/access', method: 'PUT', params: { id: '@id', rpn: '@rpn' } },
        forceUpdateService: { method: 'PUT', params: { id: '@id', action: 'forceupdateservice' } },
        registries: { url: `${API_ENDPOINT_ENDPOINTS}/:id/registries`, method: 'GET', params: { id: '@id', namespace: '@namespace' }, isArray: true },
        updateRegistryAccess: { url: `${API_ENDPOINT_ENDPOINTS}/:id/registries/:registryId`, method: 'PUT', params: { id: '@id', registryId: '@registryId' } },
      }
    );
  },
]);
