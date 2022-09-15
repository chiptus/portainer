import angular from 'angular';

import { gitCredentialsDatatable } from './GitCredentialDatatable/GitCredentialsDatatableContainer';
import { editGitCredentialViewAngular } from './EditGitCredentialView/EditGitCredentialView';
import { createGitCredentialViewAngular } from './CreateGitCredentialView/CreateGitCredentialView';

export const gitCredentialsModule = angular
  .module('portainer.app.gitCredentials', [])
  .component('gitCredentialsDatatable', gitCredentialsDatatable)
  .component('createGitCredentialViewAngular', createGitCredentialViewAngular)
  .component('editGitCredentialViewAngular', editGitCredentialViewAngular).name;
