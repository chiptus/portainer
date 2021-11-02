import angular from 'angular';
import controller from './time-window-display.controller';

angular.module('portainer.app').component('timeWindowDisplay', {
  templateUrl: './time-window-display.html',
  controller,
});
