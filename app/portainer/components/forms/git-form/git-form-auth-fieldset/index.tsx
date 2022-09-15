import { react2angular } from '@/react-tools/react2angular';

import { GitFormAuthFieldset } from './GitFormAuthFieldset';

export const GitFormAuthFieldsetReactAngular = react2angular(
  GitFormAuthFieldset,
  [
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
  ]
);
