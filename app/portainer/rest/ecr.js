angular.module('portainer.app').factory('Ecr', EcrFactory);

function EcrFactory($resource, API_ENDPOINT_REGISTRIES) {
  'use strict';
  const baseUrl = API_ENDPOINT_REGISTRIES + '/:id/ecr';

  return $resource(
    baseUrl,
    {
      id: '@id',
      repositoryName: '@repositoryName',
      tag: '@tag',
    },
    {
      deleteRepository: {
        method: 'DELETE',
        url: baseUrl + '/repositories/:repositoryName',
      },
      batchDeleteTags: {
        method: 'DELETE',
        url: baseUrl + '/repositories/:repositoryName/tags',
        headers: { 'Content-Type': 'application/json;charset=utf-8' },
        hasBody: true,
      },
    }
  );
}
