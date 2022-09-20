import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { EdgeScriptForm } from '@/react/edge/components/EdgeScriptForm';
import { withReactQuery } from '@/react-tools/withReactQuery';

import { EdgeAsyncIntervalsFormAngular } from './EdgeAsyncIntervalsForm';
import { EdgeCheckinIntervalFieldAngular } from './EdgeCheckInIntervalField';

export const componentsModule = angular
  .module('app.edge.components', [])
  .component(
    'edgeScriptForm',
    r2a(withReactQuery(EdgeScriptForm), [
      'edgeInfo',
      'commands',
      'isNomadTokenVisible',
      'hideAsyncMode',
    ])
  )
  .component('edgeCheckinIntervalField', EdgeCheckinIntervalFieldAngular)
  .component('edgeAsyncIntervalsForm', EdgeAsyncIntervalsFormAngular).name;
