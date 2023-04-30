import angular from 'angular';

import { gitCredentialsModule } from './git-credentials';

export const accountViews = angular.module(
  'portainer.app.react.views.account',
  [gitCredentialsModule]
).name;
