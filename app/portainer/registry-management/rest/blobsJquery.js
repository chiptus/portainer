/**
 * This service has been created to request the docker registry blobs API
 */

import $ from 'jquery';

angular.module('portainer.registrymanagement').factory('RegistryBlobsJquery', RegistryBlobsJqueryFactory);

function RegistryBlobsJqueryFactory(API_ENDPOINT_REGISTRIES) {
  'use strict';

  function buildUrl(params) {
    let url = API_ENDPOINT_REGISTRIES + '/' + params.id + '/v2/' + params.repository + '/blobs/' + params.digest;

    if (params.endpointId) {
      url += '?endpointId=' + params.endpointId;
    }

    return url;
  }

  function _get(params) {
    return new Promise((resolve, reject) => {
      const url = buildUrl(params);

      $.ajax({
        type: 'GET',
        url: url,
        success: (result) => resolve(result),
        error: (error) => reject(error),
      });
    });
  }

  return {
    get: _get,
  };
}
