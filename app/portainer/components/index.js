import angular from 'angular';

import dateRangePickerModule from './date-range-picker';
import gitFormModule from './forms/git-form';

export default angular.module('portainer.app.components', [dateRangePickerModule, gitFormModule]).name;
