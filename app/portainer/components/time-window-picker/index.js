import angular from 'angular';
import controller from './time-window-picker.controller';
import './time-window-picker.css';

angular.module('portainer.app').component('timeWindowPicker', {
  templateUrl: './time-window-picker.html',
  controller,
  bindings: {
    timeWindow: '=',
    timeZone: '=',
  },
});
