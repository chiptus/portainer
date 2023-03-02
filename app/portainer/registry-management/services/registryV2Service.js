import _ from 'lodash-es';
import { RegistryTypes } from 'Portainer/models/registryTypes';
import { RepositoryAddTagPayload, RepositoryShortTag } from '../models/repositoryTag';
import { RegistryRepositoryViewModel } from '../models/registryRepository';
import genericAsyncGenerator from './genericAsyncGenerator';

angular.module('portainer.registrymanagement').factory('RegistryV2Service', [
  '$q',
  '$async',
  'RegistryCatalog',
  'RegistryTags',
  'RegistryManifestsJquery',
  'RegistryV2Helper',
  function RegistryV2ServiceFactory($q, $async, RegistryCatalog, RegistryTags, RegistryManifestsJquery, RegistryV2Helper) {
    'use strict';
    var service = {};

    /**
     * PING
     */
    function ping(registry, endpointId, forceNewConfig) {
      const id = registry.Id;
      if (forceNewConfig) {
        return RegistryCatalog.pingWithForceNew({ id: id, endpointId: endpointId }).$promise;
      }
      return RegistryCatalog.ping({ id: id, endpointId: endpointId }).$promise;
    }

    /**
     * END PING
     */

    /**
     * REPOSITORIES
     */

    function _getCatalogPage(params, deferred, repositories) {
      RegistryCatalog.get(params)
        .$promise.then(function (data) {
          if (data.repositories) {
            repositories = _.concat(repositories, data.repositories);
          }
          if (data.last && data.n) {
            _getCatalogPage({ id: params.id, endpointId: params.endpointId, n: data.n, last: data.last }, deferred, repositories);
          } else {
            deferred.resolve(repositories);
          }
        })
        .catch(function error(err) {
          deferred.reject({
            msg: 'Unable to retrieve repositories',
            err: err,
          });
        });
    }

    function _getCatalog(id, endpointId) {
      var deferred = $q.defer();
      var repositories = [];

      _getCatalogPage({ id: id, endpointId: endpointId }, deferred, repositories);
      return deferred.promise;
    }

    function repositories(registry, endpointId) {
      const deferred = $q.defer();
      const id = registry.Id;

      _getCatalog(id, endpointId)
        .then(function success(data) {
          const repositories = _.map(data, (repositoryName) => new RegistryRepositoryViewModel(repositoryName));
          deferred.resolve(repositories);
        })
        .catch(function error(err) {
          deferred.reject({
            msg: 'Unable to retrieve repositories',
            err: err,
          });
        });

      return deferred.promise;
    }

    function getRepositoriesDetails(registry, endpointId, repositories) {
      const deferred = $q.defer();
      const promises = _.map(repositories, (repository) => tags(registry, endpointId, repository.Name));

      Promise.all(promises)
        .then(function success(data) {
          var repositories = data.map(function (item) {
            return new RegistryRepositoryViewModel(item);
          });
          repositories = _.without(repositories, undefined);
          deferred.resolve(repositories);
        })
        .catch(function error(err) {
          deferred.reject({
            msg: 'Unable to retrieve repositories',
            err: err,
          });
        });

      return deferred.promise;
    }

    /**
     * END REPOSITORIES
     */

    /**
     * TAGS
     */

    function _getTagsPage(params, deferred, previousTags) {
      RegistryTags.get(params)
        .$promise.then(function (data) {
          previousTags.name = data.name;
          previousTags.tags = _.concat(previousTags.tags, data.tags);
          if (data.last && data.n) {
            _getTagsPage({ id: params.id, endpointId: params.endpointId, repository: params.repository, n: data.n, last: data.last }, deferred, previousTags);
          } else {
            deferred.resolve(previousTags);
          }
        })
        .catch(function error(err) {
          if (err.status === 404) {
            deferred.resolve(previousTags);
            return;
          }

          deferred.reject({
            msg: 'Unable to retrieve tags',
            err: err,
          });
        });
    }

    function tags(registry, endpointId, repository) {
      const deferred = $q.defer();
      const id = registry.Id;

      _getTagsPage({ id: id, endpointId: endpointId, repository: repository }, deferred, { name: repository, tags: [] });
      return deferred.promise;
    }

    function getTagsDetails(registry, endpointId, repository, tags) {
      const promises = _.map(tags, (t) => tag(registry, endpointId, repository, t.Name));

      return $q.all(promises);
    }

    function tag(registry, endpointId, repository, tag) {
      const deferred = $q.defer();
      const id = registry.Id;

      var promises = {
        v1: RegistryManifestsJquery.get({
          id: id,
          endpointId: endpointId,
          repository: repository,
          tag: tag,
        }),
        v2: RegistryManifestsJquery.getV2({
          id: id,
          endpointId: endpointId,
          repository: repository,
          tag: tag,
        }),
      };
      $q.all(promises)
        .then(function success(data) {
          var tagDetails = RegistryV2Helper.manifestsToTag(data);
          tagDetails.Name = tagDetails.Name || tag;
          deferred.resolve(tagDetails);
        })
        .catch(function error(err) {
          deferred.reject({
            msg: 'Unable to retrieve tag ' + tag,
            err: err,
          });
        });

      return deferred.promise;
    }

    /**
     * END TAGS
     */

    /**
     * ADD TAG
     */

    // tag: RepositoryAddTagPayload
    function _addTagFromGenerator(registry, endpointId, repository, tag) {
      return addTag(registry, endpointId, repository, tag.Tag, tag.Manifest);
    }

    function addTag(registry, endpointId, repository, tag, manifest) {
      const id = registry.Id;
      delete manifest.digest;
      return RegistryManifestsJquery.put(
        {
          id: id,
          endpointId: endpointId,
          repository: repository,
          tag: tag,
        },
        manifest
      );
    }

    async function* _addTagsWithProgress(registry, endpointId, repository, tagsList, progression = 0) {
      // ghcr.io does not support adding multiple tags simultaneity
      const step = registry.Type === RegistryTypes.GITHUB ? 1 : undefined;
      for await (const partialResult of genericAsyncGenerator($q, tagsList, _addTagFromGenerator, [registry, endpointId, repository], step)) {
        if (typeof partialResult === 'number') {
          yield progression + partialResult;
        } else {
          yield partialResult;
        }
      }
    }

    /**
     * END ADD TAG
     */

    /**
     * DELETE MANIFEST
     */

    function deleteManifest(registry, endpointId, repository, imageDigest) {
      const id = registry.Id;
      return RegistryManifestsJquery.delete({
        id: id,
        endpointId: endpointId,
        repository: repository,
        tag: imageDigest,
      });
    }

    async function* _deleteManifestsWithProgress(registry, endpointId, repository, manifests) {
      // ghcr.io does not support deleting multiple manifests simultaneity
      const step = registry.Type === RegistryTypes.GITHUB ? 1 : undefined;
      for await (const partialResult of genericAsyncGenerator($q, manifests, deleteManifest, [registry, endpointId, repository], step)) {
        yield partialResult;
      }
    }

    /**
     * END DELETE MANIFEST
     */

    /**
     * SHORT TAG
     */

    function _shortTagFromGenerator(id, endpointId, repository, tag) {
      return new Promise((resolve, reject) => {
        RegistryManifestsJquery.getV2({ id: id, endpointId: endpointId, repository: repository, tag: tag })
          .then((data) => resolve(new RepositoryShortTag(tag, data.config.digest, data.digest, data)))
          .catch((err) => reject(err));
      });
    }

    async function* shortTagsWithProgress(registry, endpointId, repository, tagsList) {
      const id = registry.Id;
      yield* genericAsyncGenerator($q, tagsList, _shortTagFromGenerator, [id, endpointId, repository]);
    }

    /**
     * END SHORT TAG
     */

    /**
     * RETAG
     */
    async function* retagWithProgress(registry, endpointId, repository, modifiedTags, modifiedDigests, impactedTags) {
      yield* _deleteManifestsWithProgress(registry, endpointId, repository, modifiedDigests);

      const newTags = _.map(impactedTags, (item) => {
        const tagFromTable = _.find(modifiedTags, { Name: item.Name });
        const name = tagFromTable && tagFromTable.Name !== tagFromTable.NewName ? tagFromTable.NewName : item.Name;
        return new RepositoryAddTagPayload(name, item.ManifestV2);
      });

      yield* _addTagsWithProgress(registry, endpointId, repository, newTags, modifiedDigests.length);
    }

    /**
     * END RETAG
     */

    /**
     * DELETE TAGS
     */

    async function* deleteTagsWithProgress(registry, endpointId, repository, modifiedDigests, impactedTags) {
      yield* _deleteManifestsWithProgress(registry, endpointId, repository, modifiedDigests);

      const newTags = _.map(impactedTags, (item) => new RepositoryAddTagPayload(item.Name, item.ManifestV2));

      yield* _addTagsWithProgress(registry, endpointId, repository, newTags, modifiedDigests.length);
    }

    /**
     * END DELETE TAGS
     */

    /**
     * SERVICE FUNCTIONS DECLARATION
     */

    service.ping = ping;

    service.repositories = repositories;
    service.getRepositoriesDetails = getRepositoriesDetails;

    service.tags = tags;
    service.tag = tag;
    service.getTagsDetails = getTagsDetails;

    service.shortTagsWithProgress = shortTagsWithProgress;

    service.addTag = addTag;
    service.deleteManifest = deleteManifest;

    service.deleteTagsWithProgress = deleteTagsWithProgress;
    service.retagWithProgress = retagWithProgress;

    return service;
  },
]);
