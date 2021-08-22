import angular from 'angular';

import sidebarModule from './sidebar';
import dateRangePickerModule from './date-range-picker';
import gitFormModule from './forms/git-form';
import porAccessManagementModule from './accessManagement';
import formComponentsModule from './form-components';

export default angular.module('portainer.app.components', [sidebarModule, dateRangePickerModule, gitFormModule, porAccessManagementModule, formComponentsModule]).name;
