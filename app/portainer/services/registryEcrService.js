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

  function deleteRepository(registry, repository) {
    return Ecr.deleteRepository({ id: registry.Id, repositoryName: repository.Name }).$promise;
  }

  /**
   * RETAG
   */
  function _addTag(registry, repository, tag, manifest) {
    const id = registry.Id;
    delete manifest.digest;
    return RegistryManifestsJquery.put(
      {
        id: id,
        repository: repository,
        tag: tag,
      },
      manifest
    );
  }

  function _addTagFromGenerator(registry, repository, tag) {
    return _addTag(registry, repository, tag.NewName, tag.ManifestV2);
  }

  async function* _addTagsWithProgress(registry, repository, modifiedTags) {
    for await (const partialResult of genericAsyncGenerator($q, modifiedTags, _addTagFromGenerator, [registry, repository])) {
      yield partialResult;
    }
  }

  async function* retagWithProgress(registry, repository, modifiedTags, modifiedDigests, impactedTags) {
    yield* _addTagsWithProgress(registry, repository, modifiedTags);

    const oldTagNames = modifiedTags.map((tag) => tag.Name);
    const newTagNames = modifiedTags.map((tag) => tag.NewName);
    const toDelTags = _.without(oldTagNames, ...newTagNames);

    if (toDelTags.length) {
      await batchDeleteTags({ id: registry.Id, repositoryName: repository }, { Tags: toDelTags });
    }

    yield modifiedDigests.length + impactedTags.length;
  }

  return service;
}
