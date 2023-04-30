import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { OpenAIKeyWidget } from '@/react/portainer/account/AccountView/OpenAIKey';

export const accountModule = angular
  .module('portainer.app.react.components.account', [])

  .component(
    'openAiKeyWidget',
    r2a(withUIRouter(withReactQuery(withCurrentUser(OpenAIKeyWidget))), [])
  ).name;
