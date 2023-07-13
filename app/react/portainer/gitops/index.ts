import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { PathSelector } from '@/react/portainer/gitops/ComposePathField/PathSelector';

export const ngModule = angular
  .module('portainer.app.react.gitops', [])
  .component(
    'pathSelector',
    r2a(withUIRouter(withReactQuery(PathSelector)), [
      'value',
      'onChange',
      'placeholder',
      'model',
      'dirOnly',
    ])
  );

export const gitopsModule = ngModule.name;
