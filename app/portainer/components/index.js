import angular from 'angular';

import dateRangePickerModule from './date-range-picker';
import gitFormModule from './forms/git-form';
import porAccessManagementModule from './accessManagement';
import formComponentsModule from './form-components';

export default angular.module('portainer.app.components', [dateRangePickerModule, gitFormModule, porAccessManagementModule, formComponentsModule]).name;
