import angular from 'angular';

import sidebarModule from './sidebar';
import dateRangePickerModule from './date-range-picker';
import gitFormModule from './forms/git-form';
import porAccessManagementModule from './accessManagement';
import formComponentsModule from './form-components';
import widgetModule from './widget';

import { ReactExampleAngular } from './ReactExample';
import { TooltipAngular } from './Tooltip';

export default angular
  .module('portainer.app.components', [widgetModule, sidebarModule, dateRangePickerModule, gitFormModule, porAccessManagementModule, formComponentsModule])
  .component('portainerTooltip', TooltipAngular)
  .component('reactExample', ReactExampleAngular).name;
