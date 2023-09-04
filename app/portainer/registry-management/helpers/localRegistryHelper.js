import _ from 'lodash-es';
import { RepositoryTagViewModel } from '../models/repositoryTag';

angular.module('portainer.registrymanagement').factory('RegistryV2Helper', [
  function RegistryV2HelperFactory() {
    'use strict';

    var helper = {};

    function parseV1History(history) {
      return _.map(history, (item) => angular.fromJson(item.v1Compatibility));
    }

    // convert image configs blob history to manifest v1 history
    function parseImageConfigsHistory(imageConfigs, v2) {
      return _.map(imageConfigs.history.reverse(), (item) => {
        item.CreatedBy = item.created_by;

        // below fields exist in manifest v1 history but not image configs blob
        item.id = v2.config.digest;
        item.created = imageConfigs.created;
        item.docker_version = imageConfigs.docker_version;
        item.os = imageConfigs.os;
        item.architecture = imageConfigs.architecture;
        item.config = imageConfigs.config;
        item.container_config = imageConfigs.container_config;

        return item;
      });
    }

    helper.manifestsToTag = function (manifests) {
      var v1 = manifests.v1;
      var v2 = manifests.v2;
      var imageConfigs = manifests.imageConfigs;

      var history = [];
      var name = '';
      var os = '';
      var arch = '';

      if (imageConfigs) {
        // use info from image configs blob when manifest v1 is not provided by registry
        os = imageConfigs.os || '';
        arch = imageConfigs.architecture || '';
        history = parseImageConfigsHistory(imageConfigs, v2);
      } else if (v1) {
        // use info from manifest v1
        history = parseV1History(v1.history);
        name = v1.tag;
        os = _.get(history, '[0].os', '');
        arch = v1.architecture;
      }

      var size = v2.layers.reduce(function (a, b) {
        return {
          size: a.size + b.size,
        };
      }).size;

      var imageId = v2.config.digest;

      // v2.digest comes from
      //  1. Docker-Content-Digest header from the v2 response, or
      //  2. Calculated locally by sha256(v2-response-body)
      var imageDigest = v2.digest;

      return new RepositoryTagViewModel(name, os, arch, size, imageDigest, imageId, v2, history);
    };

    return helper;
  },
]);
