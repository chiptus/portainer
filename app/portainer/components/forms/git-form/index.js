import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { GitFormComposePathField } from '@/react/portainer/gitops/GitFormComposePathField';
import { GitFormUrlField } from '@/react/portainer/gitops/GitFormUrlField';
import { GitFormRefField } from '@/react/portainer/gitops/GitFormRefField';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { GitFormAuthFieldset } from '@/portainer/components/forms/git-form/git-form-auth-fieldset/GitFormAuthFieldset';
import { gitForm } from './git-form';
import { gitFormAdditionalFilesPanel } from './git-form-additional-files-panel';
import { gitFormAdditionalFileItem } from './/git-form-additional-files-panel/git-form-additional-file-item';
import { gitFormAutoUpdateFieldset } from './git-form-auto-update-fieldset';
import { gitFormInfoPanel } from './git-form-info-panel';

export default angular
  .module('portainer.app.components.forms.git', [])
  .component('gitForm', gitForm)
  .component('gitFormInfoPanel', gitFormInfoPanel)
  .component('gitFormAdditionalFilesPanel', gitFormAdditionalFilesPanel)
  .component('gitFormAdditionalFileItem', gitFormAdditionalFileItem)
  .component('gitFormAutoUpdateFieldset', gitFormAutoUpdateFieldset)
  .component('gitFormComposePathField', r2a(withUIRouter(withReactQuery(withCurrentUser(GitFormComposePathField))), ['value', 'onChange', 'isCompose', 'model']))
  .component('gitFormRefField', r2a(withUIRouter(withReactQuery(withCurrentUser(GitFormRefField))), ['value', 'onChange', 'model']))
  .component(
    'gitFormUrlField',
    r2a(withUIRouter(withReactQuery(withCurrentUser(GitFormUrlField))), ['value', 'onChange', 'onChangeRepositoryValid', 'onRefreshGitopsCache', 'model'])
  )
  .component(
    'gitFormAuthFieldset',
    r2a(withUIRouter(withReactQuery(withCurrentUser(GitFormAuthFieldset))), [
      'repositoryAuthentication',
      'repositoryUsername',
      'repositoryPassword',
      'gitCredentials',
      'selectedGitCredential',
      'saveCredential',
      'showAuthExplanation',
      'newCredentialName',
      'newCredentialNameExist',
      'newCredentialNameInvalid',
      'onSelectGitCredential',
      'onChangeRepositoryAuthentication',
      'onChangeRepositoryUsername',
      'onChangeRepositoryPassword',
      'onChangeSaveCredential',
      'onChangeNewCredentialName',
    ])
  ).name;
