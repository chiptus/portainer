import angular from 'angular';

angular.module('portainer.kubernetes').factory('Opa', OpaFactory);

/* @ngInject */
function OpaFactory($resource, API_ENDPOINT_KUBERNETES, EndpointProvider) {
  const url = API_ENDPOINT_KUBERNETES + '/:endpointId/opa';
  return $resource(
    url,
    {
      endpointId: EndpointProvider.endpointID,
    },
    {
      fetch: {
        method: 'GET',
        ignoreLoadingBar: true,
        transformResponse: (data) => ({ data: JSON.parse(data) }),
      },
      update: {
        method: 'put',
        ignoreLoadingBar: true,
        transformResponse: (data) => ({ data: JSON.parse(data) }),
      },
    }
  );
}
