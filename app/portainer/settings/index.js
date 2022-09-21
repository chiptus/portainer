import angular from 'angular';

import generalModule from './general';
import authenticationModule from './authentication';

export default angular.module('portainer.settings', [authenticationModule, generalModule]).name;
