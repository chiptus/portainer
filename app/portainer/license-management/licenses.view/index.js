import angular from 'angular';

import { licensesView } from './licenses.view';
import { licensesDatatable } from './licenses-datatable';

export default angular.module('portainer.app.license-management.licenses-view', []).component('licensesDatatable', licensesDatatable).component('licensesView', licensesView).name;
