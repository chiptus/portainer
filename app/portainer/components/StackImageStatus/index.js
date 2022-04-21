import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { StackImageStatus } from 'Portainer/components/StackImageStatus/StackImageStatus';

const ImageStatusAngular = r2a(StackImageStatus, ['stackId']);

export default angular.module('app.portainer.component.stack-image-status', []).component('stackImageStatus', ImageStatusAngular).name;
