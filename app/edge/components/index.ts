import angular from 'angular';

import { EdgeAsyncIntervalsFormAngular } from './EdgeAsyncIntervalsForm';
import { EdgeCheckinIntervalFieldAngular } from './EdgeCheckInIntervalField';
import { EdgeScriptFormAngular } from './EdgeScriptForm';

export const componentsModule = angular
  .module('app.edge.components', [])
  .component('edgeScriptForm', EdgeScriptFormAngular)
  .component('edgeCheckinIntervalField', EdgeCheckinIntervalFieldAngular)
  .component('edgeAsyncIntervalsForm', EdgeAsyncIntervalsFormAngular).name;
