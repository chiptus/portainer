import angular from 'angular';

angular.module('portainer.edge').factory('EdgeStackService', function EdgeStackServiceFactory(EdgeStacks, FileUploadService) {
  var service = {};

  service.stack = function stack(id) {
    return EdgeStacks.get({ id }).$promise;
  };

  service.stacks = function stacks() {
    return EdgeStacks.query({}).$promise;
  };

  service.remove = function remove(id) {
    return EdgeStacks.remove({ id }).$promise;
  };

  service.stackFile = async function stackFile(id) {
    try {
      const { StackFileContent } = await EdgeStacks.file({ id }).$promise;
      return StackFileContent;
    } catch (err) {
      throw { msg: 'Unable to retrieve stack content', err };
    }
  };

  service.updateStack = async function updateStack(id, stack, registries) {
    return EdgeStacks.update({ id }, stack, registries).$promise;
  };

  service.createStackFromFileContent = async function createStackFromFileContent(payload, dryrun) {
    try {
      return await EdgeStacks.create({ method: 'string', dryrun: dryrun ? 'true' : 'false' }, payload).$promise;
    } catch (err) {
      throw { msg: 'Unable to create the stack', err };
    }
  };

  service.createStackFromFileUpload = async function createStackFromFileUpload(payload, file, dryrun) {
    try {
      return await FileUploadService.createEdgeStack(payload, file, dryrun);
    } catch (err) {
      throw { msg: 'Unable to create the stack', err };
    }
  };

  service.createStackFromGitRepository = async function createStackFromGitRepository(payload, repositoryOptions) {
    try {
      return await EdgeStacks.create(
        { method: 'repository' },
        {
          ...payload,
          RepositoryURL: repositoryOptions.RepositoryURL,
          RepositoryReferenceName: repositoryOptions.RepositoryReferenceName,
          FilePathInRepository: repositoryOptions.FilePathInRepository,
          RepositoryAuthentication: repositoryOptions.RepositoryAuthentication,
          RepositoryUsername: repositoryOptions.RepositoryUsername,
          RepositoryPassword: repositoryOptions.RepositoryPassword,
        }
      ).$promise;
    } catch (err) {
      throw { msg: 'Unable to create the stack', err };
    }
  };

  service.update = function update(stack) {
    return EdgeStacks.update(stack).$promise;
  };

  return service;
});
