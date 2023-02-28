import angular from 'angular';

import { limitedFeatureDirective } from './limited-feature.directive';

import './feature-flags.css';

export const featureFlagsModule = angular
  .module('portainer.feature-flags', [])
  .directive('limitedFeatureDir', limitedFeatureDirective).name;
