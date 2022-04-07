import { RegistryTypes } from 'Portainer/models/registryTypes';

angular.module('portainer.registrymanagement').factory('RegistryServiceSelector', [
  '$q',
  'RegistryV2Service',
  'RegistryGitlabService',
  'RegistryEcrService',
  function RegistryServiceSelector($q, RegistryV2Service, RegistryGitlabService, RegistryEcrService) {
    'use strict';
    return {
      ping,
      repositories,
      getRepositoriesDetails,
      tag,
      tags,
      getTagsDetails,
      addTag,
      retagWithProgress,
      shortTagsWithProgress,
      deleteTagsWithProgress,
      deleteManifest,
      deleteRepository,
      batchDeleteTags,
    };

    function ping(registry, endpointId, forceNewConfig) {
      return RegistryV2Service.ping(registry, endpointId, forceNewConfig);
    }

    function repositories(registry, endpointId) {
      if (registry.Type === RegistryTypes.GITLAB) {
        return RegistryGitlabService.repositories(registry, endpointId);
      }
      return RegistryV2Service.repositories(registry, endpointId);
    }

    function getRepositoriesDetails(registry, endpointId, repositories) {
      return RegistryV2Service.getRepositoriesDetails(registry, endpointId, repositories);
    }

    function tags(registry, endpointId, repository) {
      return RegistryV2Service.tags(registry, endpointId, repository);
    }

    function getTagsDetails(registry, endpointId, repository, tags) {
      return RegistryV2Service.getTagsDetails(registry, endpointId, repository, tags);
    }

    function tag(registry, endpointId, repository, tag) {
      return RegistryV2Service.tag(registry, endpointId, repository, tag);
    }

    function addTag(registry, endpointId, repository, tag, manifest) {
      return RegistryV2Service.addTag(registry, endpointId, repository, tag, manifest);
    }

    function deleteManifest(registry, endpointId, repository, digest) {
      return RegistryV2Service.deleteManifest(registry, endpointId, repository, digest);
    }

    function shortTagsWithProgress(registry, endpointId, repository, tagsList) {
      return RegistryV2Service.shortTagsWithProgress(registry, endpointId, repository, tagsList);
    }

    function deleteTagsWithProgress(registry, endpointId, repository, modifiedDigests, impactedTags) {
      return RegistryV2Service.deleteTagsWithProgress(registry, endpointId, repository, modifiedDigests, impactedTags);
    }

    function retagWithProgress(registry, endpointId, repository, modifiedTags, modifiedDigests, impactedTags) {
      if (registry.Type === RegistryTypes.ECR) {
        return RegistryEcrService.retagWithProgress(registry, endpointId, repository, modifiedTags, modifiedDigests, impactedTags);
      }
      return RegistryV2Service.retagWithProgress(registry, endpointId, repository, modifiedTags, modifiedDigests, impactedTags);
    }

    // only for ECR
    function batchDeleteTags(params, data) {
      return RegistryEcrService.batchDeleteTags(params, data);
    }

    // only for ECR
    function deleteRepository(registry, endpointId, repository) {
      return RegistryEcrService.deleteRepository(registry, endpointId, repository);
    }
  },
]);
