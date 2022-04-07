import _ from 'lodash-es';
import genericAsyncGenerator from '@/portainer/registry-management/services/genericAsyncGenerator';

angular.module('portainer.app').factory('RegistryEcrService', RegistryEcrServiceFactory);

/* @ngInject */
function RegistryEcrServiceFactory($q, RegistryManifestsJquery, Ecr) {
  'use strict';
  const service = {
    retagWithProgress,
    batchDeleteTags,
    deleteRepository,
  };

  function batchDeleteTags(params, data) {
    return Ecr.batchDeleteTags(params, data).$promise;
  }

  function deleteRepository(registry, endpointId, repository) {
    return Ecr.deleteRepository({ id: registry.Id, endpointId: endpointId, repositoryName: repository.Name }).$promise;
  }

  /**
   * RETAG
   */
  function _addTag(registry, endpointId, repository, tag, manifest) {
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

  function _addTagFromGenerator(registry, endpointId, repository, tag) {
    return _addTag(registry, endpointId, repository, tag.NewName, tag.ManifestV2);
  }

  async function* _addTagsWithProgress(registry, endpointId, repository, modifiedTags) {
    for await (const partialResult of genericAsyncGenerator($q, modifiedTags, _addTagFromGenerator, [registry, endpointId, repository])) {
      yield partialResult;
    }
  }

  async function* retagWithProgress(registry, endpointId, repository, modifiedTags, modifiedDigests, impactedTags) {
    yield* _addTagsWithProgress(registry, endpointId, repository, modifiedTags);

    const oldTagNames = modifiedTags.map((tag) => tag.Name);
    const newTagNames = modifiedTags.map((tag) => tag.NewName);
    const toDelTags = _.without(oldTagNames, ...newTagNames);

    if (toDelTags.length) {
      await batchDeleteTags({ id: registry.Id, endpointId: endpointId, repositoryName: repository }, { Tags: toDelTags });
    }

    yield modifiedDigests.length + impactedTags.length;
  }

  return service;
}
