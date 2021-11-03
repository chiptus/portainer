import angular from 'angular';

import sidebarModule from './sidebar';
import dateRangePickerModule from './date-range-picker';
import gitFormModule from './forms/git-form';
import porAccessManagementModule from './accessManagement';
import formComponentsModule from './form-components';

import { ReactExampleAngular } from './ReactExample';

export default angular
  .module('portainer.app.components', [sidebarModule, dateRangePickerModule, gitFormModule, porAccessManagementModule, formComponentsModule])
  .component('reactExample', ReactExampleAngular).name;
