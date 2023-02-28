import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { EditGitCredentialsView } from '@/react/portainer/account/git-credentials/EditGitCredentialsView';
import { CreateGitCredentialsView } from '@/react/portainer/account/git-credentials/CreateGitCredentialsView';

export const gitCredentialsModule = angular
  .module('portainer.app.react.views.account.git-credentials', [])

  .component(
    'createGitCredentialView',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(CreateGitCredentialsView))),
      []
    )
  )
  .component(
    'editGitCredentialView',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(EditGitCredentialsView))),
      []
    )
  ).name;
