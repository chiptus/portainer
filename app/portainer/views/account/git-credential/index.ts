import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { withReactQuery } from '@/react-tools/withReactQuery';

import EditGitCredentialView from './EditGitCredentialView/EditGitCredentialView';
import CreateGitCredentialView from './CreateGitCredentialView/CreateGitCredentialView';
import { GitCredentialsDatatable } from './GitCredentialDatatable';

export const gitCredentialsModule = angular
  .module('portainer.account.git', [])
  .component(
    'gitCredentialsDatatable',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(GitCredentialsDatatable))),
      []
    )
  )
  .component(
    'createGitCredentialViewAngular',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(CreateGitCredentialView))),
      []
    )
  )
  .component(
    'editGitCredentialViewAngular',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(EditGitCredentialView))),
      []
    )
  ).name;
