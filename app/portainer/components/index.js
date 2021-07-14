import angular from 'angular';

import dateRangePickerModule from './date-range-picker';
import gitFormModule from './forms/git-form';
import porAccessManagementModule from './accessManagement';

export default angular.module('portainer.app.components', [dateRangePickerModule, gitFormModule, porAccessManagementModule]).name;
