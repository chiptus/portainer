angular.module('portainer.app').factory('OpaService', [
  '$q',
  '$async',
  'Opa',
  function OpaServiceFactory($q, $async, Opa) {
    'use strict';
    const service = {};

    service.detail = function () {
      const deferred = $q.defer();

      Opa.fetch()
        .$promise.then(function success(data) {
          deferred.resolve(data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve opa details', err: err });
        });

      return deferred.promise;
    };

    service.save = function (form) {
      const deferred = $q.defer();

      Opa.update(form)
        .$promise.then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to save constraint settings', err: err });
        });

      return deferred.promise;
    };

    return service;
  },
]);
