import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';

import { ImageStatus } from './ImageStatus';

export { ImageStatus };

const ImageStatusAngular = r2a(ImageStatus, ['imageName', 'environmentId']);

export default angular.module('app.docker.component.image-status', []).component('imageStatus', ImageStatusAngular).name;
