import _ from 'lodash-es';
import { transformAutoUpdateViewModel } from '@/react/portainer/gitops/AutoUpdateFieldset/utils';
import { StackViewModel } from '@/react/docker/stacks/view-models/stack';

angular.module('portainer.app').factory('StackService', [
  '$q',
  '$async',
  'Stack',
  'FileUploadService',
  'StackHelper',
  'ServiceService',
  'ContainerService',
  'SwarmService',
  function StackServiceFactory($q, $async, Stack, FileUploadService, StackHelper, ServiceService, ContainerService, SwarmService) {
    'use strict';
    var service = {
      updateGit,
      updateKubeGit,
    };

    service.stack = function (id) {
      var deferred = $q.defer();

      Stack.get({ id: id })
        .$promise.then(function success(data) {
          var stack = new StackViewModel(data);
          deferred.resolve(stack);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stack details', err: err });
        });

      return deferred.promise;
    };

    service.getStackFile = function (id, version, commitHash) {
      var deferred = $q.defer();

      Stack.getStackFile({ id: id, version: version, commitHash: commitHash })
        .$promise.then(function success(data) {
          deferred.resolve(data.StackFileContent);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stack content', err: err });
        });

      return deferred.promise;
    };

    service.migrateSwarmStack = function (stack, targetEndpointId, newName) {
      var deferred = $q.defer();

      SwarmService.swarm(targetEndpointId)
        .then(function success(data) {
          var swarm = data;
          if (swarm.Id === stack.SwarmId) {
            deferred.reject({ msg: 'Target environment is located in the same Swarm cluster as the current environment', err: null });
            return;
          }
          return Stack.migrate({ id: stack.Id, endpointId: stack.EndpointId }, { EndpointID: targetEndpointId, SwarmID: swarm.Id, Name: newName }).$promise;
        })
        .then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to migrate stack', err: err });
        });

      return deferred.promise;
    };

    service.migrateComposeStack = function (stack, targetEndpointId, newName) {
      var deferred = $q.defer();

      Stack.migrate({ id: stack.Id, endpointId: stack.EndpointId }, { EndpointID: targetEndpointId, Name: newName })
        .$promise.then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to migrate stack', err: err });
        });

      return deferred.promise;
    };

    service.stacks = function (compose, swarm, endpointId, includeOrphanedStacks = false) {
      var deferred = $q.defer();

      var queries = [];
      if (compose) {
        queries.push(service.composeStacks(endpointId, true, { EndpointID: endpointId, IncludeOrphanedStacks: includeOrphanedStacks }));
      }
      if (swarm) {
        queries.push(service.swarmStacks(endpointId, true, { EndpointID: endpointId, IncludeOrphanedStacks: includeOrphanedStacks }));
      }

      $q.all(queries)
        .then(function success(data) {
          var stacks = [];
          if (data[0]) {
            stacks = stacks.concat(data[0]);
          }
          if (data[1]) {
            stacks = stacks.concat(data[1]);
          }
          stacks = _.uniqWith(stacks, (s1, s2) => s1.Id === s2.Id && s1.Name === s2.Name);
          deferred.resolve(stacks);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stacks', err: err });
        });

      return deferred.promise;
    };

    service.externalSwarmStacks = function () {
      var deferred = $q.defer();

      ServiceService.services()
        .then(function success(services) {
          deferred.resolve(StackHelper.getExternalStacksFromServices(services));
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve external stacks', err: err });
        });

      return deferred.promise;
    };

    service.externalComposeStacks = function () {
      var deferred = $q.defer();

      ContainerService.containers(1)
        .then(function success(containers) {
          deferred.resolve(StackHelper.getExternalStacksFromContainers(containers));
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve external stacks', err: err });
        });

      return deferred.promise;
    };

    service.unionStacks = function (stacks, externalStacks) {
      stacks.forEach((stack) => {
        externalStacks.forEach((externalStack) => {
          if (stack.Orphaned && stack.Name == externalStack.Name) {
            stack.OrphanedRunning = true;
          }
        });
      });
      const result = _.unionWith(stacks, externalStacks, function (a, b) {
        return a.Name === b.Name;
      });

      return result;
    };

    service.composeStacks = function (endpointId, includeExternalStacks, filters) {
      var deferred = $q.defer();

      $q.all({
        stacks: Stack.query({ filters: filters }).$promise,
        externalStacks: includeExternalStacks ? service.externalComposeStacks() : [],
      })
        .then(function success(data) {
          var stacks = data.stacks.map(function (item) {
            return new StackViewModel(item, item.EndpointId != endpointId);
          });
          var externalStacks = data.externalStacks;
          const result = service.unionStacks(stacks, externalStacks);
          deferred.resolve(result);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stacks', err: err });
        });

      return deferred.promise;
    };

    service.swarmStacks = function (endpointId, includeExternalStacks, filters = {}) {
      var deferred = $q.defer();

      SwarmService.swarm()
        .then(function success(data) {
          var swarm = data;
          filters = { SwarmID: swarm.Id, ...filters };

          return $q.all({
            stacks: Stack.query({ filters: filters }).$promise,
            externalStacks: includeExternalStacks ? service.externalSwarmStacks() : [],
          });
        })
        .then(function success(data) {
          var stacks = data.stacks.map(function (item) {
            return new StackViewModel(item, item.EndpointId != endpointId);
          });
          var externalStacks = data.externalStacks;
          const result = service.unionStacks(stacks, externalStacks);
          deferred.resolve(result);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stacks', err: err });
        });

      return deferred.promise;
    };

    service.remove = function (stack, external, endpointId) {
      var deferred = $q.defer();

      Stack.remove({ id: stack.Id ? stack.Id : stack.Name, external: external, endpointId: endpointId })
        .$promise.then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to remove the stack', err: err });
        });

      return deferred.promise;
    };

    service.associate = function (stack, endpointId, orphanedRunning) {
      var deferred = $q.defer();

      if (stack.Type == 1) {
        SwarmService.swarm()
          .then(function success(data) {
            const swarm = data;
            return Stack.associate({ id: stack.Id, endpointId: endpointId, swarmId: swarm.Id, orphanedRunning }).$promise;
          })
          .then(function success(data) {
            deferred.resolve(data);
          })
          .catch(function error(err) {
            deferred.reject({ msg: 'Unable to associate the stack', err: err });
          });
      } else {
        Stack.associate({ id: stack.Id, endpointId: endpointId, orphanedRunning })
          .$promise.then(function success(data) {
            deferred.resolve(data);
          })
          .catch(function error(err) {
            deferred.reject({ msg: 'Unable to associate the stack', err: err });
          });
      }

      return deferred.promise;
    };

    service.updateStack = function (stack, stackFile, env, prune, webhook, pullImage, rollbackTo) {
      return Stack.update(
        { endpointId: stack.EndpointId },
        {
          id: stack.Id,
          StackFileContent: stackFile,
          Env: env,
          Prune: prune,
          Webhook: webhook,
          PullImage: pullImage,
          RollbackTo: rollbackTo,
        }
      ).$promise;
    };

    service.updateKubeStack = function (stack, { stackFile, gitConfig, webhookId }) {
      let payload = {};

      if (stackFile) {
        payload = {
          StackFileContent: stackFile,
        };
      } else {
        payload = {
          AutoUpdate: transformAutoUpdateViewModel(gitConfig.AutoUpdate, webhookId),
          RepositoryReferenceName: gitConfig.RefName,
          RepositoryAuthentication: gitConfig.RepositoryAuthentication,
          RepositoryUsername: gitConfig.RepositoryUsername,
          RepositoryPassword: gitConfig.RepositoryPassword,
          RepositoryGitCredentialID: gitConfig.RepositoryGitCredentialID,
          TLSSkipVerify: gitConfig.TLSSkipVerify,
        };
      }

      return Stack.update({ id: stack.Id, endpointId: stack.EndpointId }, payload).$promise;
    };

    service.createComposeStackFromFileUpload = function (name, stackFile, env, endpointId, webhook) {
      return FileUploadService.createComposeStack(name, stackFile, env, endpointId, webhook);
    };

    service.createSwarmStackFromFileUpload = function (name, stackFile, env, endpointId, webhook) {
      var deferred = $q.defer();

      SwarmService.swarm()
        .then(function success(data) {
          var swarm = data;
          return FileUploadService.createSwarmStack(name, swarm.Id, stackFile, env, endpointId, webhook);
        })
        .then(function success(data) {
          deferred.resolve(data.data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to create the stack', err: err });
        });

      return deferred.promise;
    };
    service.createComposeStackFromFileContent = function (name, stackFileContent, env, endpointId, webhook) {
      var payload = {
        Name: name,
        StackFileContent: stackFileContent,
        Env: env,
        Webhook: webhook,
      };
      return Stack.create({ endpointId: endpointId }, { method: 'string', type: 'standalone', ...payload }).$promise;
    };

    service.createSwarmStackFromFileContent = function (name, stackFileContent, env, endpointId, webhook) {
      var deferred = $q.defer();

      SwarmService.swarm(endpointId)
        .then(function success(swarm) {
          var payload = {
            Name: name,
            SwarmID: swarm.Id,
            StackFileContent: stackFileContent,
            Env: env,
            Webhook: webhook,
          };
          return Stack.create({ endpointId: endpointId }, { method: 'string', type: 'swarm', ...payload }).$promise;
        })
        .then(function success(data) {
          deferred.resolve(data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to create the stack', err: err });
        });

      return deferred.promise;
    };

    service.createComposeStackFromGitRepository = function (name, repositoryOptions, env, endpointId) {
      var payload = {
        Name: name,
        RepositoryURL: repositoryOptions.RepositoryURL,
        RepositoryReferenceName: repositoryOptions.RepositoryReferenceName,
        ComposeFile: repositoryOptions.ComposeFilePathInRepository,
        AdditionalFiles: repositoryOptions.AdditionalFiles,
        RepositoryAuthentication: repositoryOptions.RepositoryAuthentication,
        RepositoryUsername: repositoryOptions.RepositoryUsername,
        RepositoryPassword: repositoryOptions.RepositoryPassword,
        RepositoryGitCredentialID: repositoryOptions.RepositoryGitCredentialID,
        FromAppTemplate: repositoryOptions.FromAppTemplate,
        Env: env,
        SupportRelativePath: repositoryOptions.SupportRelativePath,
        FilesystemPath: repositoryOptions.FilesystemPath,
        TLSSkipVerify: repositoryOptions.TLSSkipVerify,
      };

      if (repositoryOptions.AutoUpdate) {
        payload.AutoUpdate = repositoryOptions.AutoUpdate;
      }

      return Stack.create({ endpointId: endpointId }, { method: 'repository', type: 'standalone', ...payload }).$promise;
    };

    service.createSwarmStackFromGitRepository = function (name, repositoryOptions, env, endpointId) {
      var deferred = $q.defer();

      SwarmService.swarm()
        .then(function success(data) {
          var swarm = data;
          var payload = {
            Name: name,
            SwarmID: swarm.Id,
            RepositoryURL: repositoryOptions.RepositoryURL,
            RepositoryReferenceName: repositoryOptions.RepositoryReferenceName,
            ComposeFile: repositoryOptions.ComposeFilePathInRepository,
            AdditionalFiles: repositoryOptions.AdditionalFiles,
            RepositoryAuthentication: repositoryOptions.RepositoryAuthentication,
            RepositoryUsername: repositoryOptions.RepositoryUsername,
            RepositoryPassword: repositoryOptions.RepositoryPassword,
            RepositoryGitCredentialID: repositoryOptions.RepositoryGitCredentialID,
            FromAppTemplate: repositoryOptions.FromAppTemplate,
            Env: env,
            SupportRelativePath: repositoryOptions.SupportRelativePath,
            FilesystemPath: repositoryOptions.FilesystemPath,
            TLSSkipVerify: repositoryOptions.TLSSkipVerify,
          };

          if (repositoryOptions.AutoUpdate) {
            payload.AutoUpdate = repositoryOptions.AutoUpdate;
          }

          return Stack.create({ endpointId: endpointId }, { method: 'repository', type: 'swarm', ...payload }).$promise;
        })
        .then(function success(data) {
          deferred.resolve(data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to create the stack', err: err });
        });

      return deferred.promise;
    };

    service.duplicateStack = function duplicateStack(name, stackFileContent, env, endpointId, type) {
      var action = type === 1 ? service.createSwarmStackFromFileContent : service.createComposeStackFromFileContent;
      return action(name, stackFileContent, env, endpointId);
    };

    async function kubernetesDeployAsync(endpointId, method, payload) {
      try {
        await Stack.create({ endpointId: endpointId }, { method, type: 'kubernetes', ...payload }).$promise;
      } catch (err) {
        throw { err: err };
      }
    }

    service.kubernetesDeploy = function (endpointId, method, payload) {
      return $async(kubernetesDeployAsync, endpointId, method, payload);
    };

    service.start = start;
    function start(endpointId, id) {
      return Stack.start({ id, endpointId }).$promise;
    }

    service.stop = stop;
    function stop(endpointId, id) {
      return Stack.stop({ endpointId, id }).$promise;
    }

    function updateGit(id, endpointId, env, prune, gitConfig, pullImage) {
      return Stack.updateGit(
        { endpointId, id },
        {
          env,
          prune,
          RepositoryReferenceName: gitConfig.RefName,
          RepositoryAuthentication: gitConfig.RepositoryAuthentication,
          RepositoryGitCredentialID: gitConfig.RepositoryGitCredentialID,
          RepositoryUsername: gitConfig.RepositoryUsername,
          RepositoryPassword: gitConfig.RepositoryPassword,
          PullImage: pullImage,
        }
      ).$promise;
    }

    function updateKubeGit(id, endpointId, namespace, gitConfig) {
      return Stack.updateGit(
        { endpointId, id },
        {
          Namespace: namespace,
          RepositoryReferenceName: gitConfig.RefName,
          RepositoryAuthentication: gitConfig.RepositoryAuthentication,
          RepositoryUsername: gitConfig.RepositoryUsername,
          RepositoryPassword: gitConfig.RepositoryPassword,
          RepositoryGitCredentialID: gitConfig.RepositoryGitCredentialID,
        }
      ).$promise;
    }

    service.updateGitStackSettings = function (id, endpointId, env, gitConfig, webhookId) {
      return Stack.updateGitStackSettings(
        { endpointId, id },
        {
          AutoUpdate: transformAutoUpdateViewModel(gitConfig.AutoUpdate, webhookId),
          Env: env,
          RepositoryReferenceName: gitConfig.RefName,
          RepositoryAuthentication: gitConfig.RepositoryAuthentication,
          RepositoryGitCredentialID: gitConfig.RepositoryGitCredentialID,
          RepositoryUsername: gitConfig.RepositoryUsername,
          RepositoryPassword: gitConfig.RepositoryPassword,
          Prune: gitConfig.Option.Prune,
          TLSSkipVerify: gitConfig.TLSSkipVerify,
        }
      ).$promise;
    };

    return service;
  },
]);
