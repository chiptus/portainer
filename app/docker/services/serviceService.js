import { ServiceViewModel } from '../models/service';

angular.module('portainer.docker').factory('ServiceService', [
  '$q',
  'Service',
  function ServiceServiceFactory($q, Service) {
    'use strict';
    var service = {};

    service.services = function (filters) {
      var deferred = $q.defer();

      Service.query({ filters: filters ? filters : {} })
        .$promise.then(function success(data) {
          var services = data.map(function (item) {
            return new ServiceViewModel(item);
          });
          deferred.resolve(services);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve services', err: err });
        });

      return deferred.promise;
    };

    service.service = function (id) {
      var deferred = $q.defer();

      Service.get({ id: id })
        .$promise.then(function success(data) {
          var service = new ServiceViewModel(data);
          deferred.resolve(service);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve service details', err: err });
        });

      return deferred.promise;
    };

    service.remove = function (service) {
      var deferred = $q.defer();

      Service.remove({ id: service.Id })
        .$promise.then(function success(data) {
          if (data.message) {
            deferred.reject({ msg: data.message, err: data.message });
          } else {
            deferred.resolve();
          }
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to remove service', err: err });
        });

      return deferred.promise;
    };

    service.update = function (serv, config, rollback) {
      return service.service(serv.Id).then((data) => {
        const params = {
          id: serv.Id,
          version: data.Version,
        };
        if (rollback) {
          params.rollback = rollback;
        }
        return Service.update(params, config).$promise;
      });
    };

    return service;
  },
]);
