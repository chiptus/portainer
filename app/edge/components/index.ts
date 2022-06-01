import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { EdgeScriptForm } from '@/react/edge/components/EdgeScriptForm';

import { EdgeAsyncIntervalsFormAngular } from './EdgeAsyncIntervalsForm';
import { EdgeCheckinIntervalFieldAngular } from './EdgeCheckInIntervalField';

export const componentsModule = angular
  .module('app.edge.components', [])
  .component(
    'edgeScriptForm',
    r2a(EdgeScriptForm, ['edgeInfo', 'commands', 'isNomadTokenVisible'])
  )
  .component('edgeCheckinIntervalField', EdgeCheckinIntervalFieldAngular)
  .component('edgeAsyncIntervalsForm', EdgeAsyncIntervalsFormAngular).name;
