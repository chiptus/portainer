import angular from 'angular';

import { LicenseInfoPanelAngular } from '@@/LicenseInfoPanel/LicenseInfoPanel';
import { licensesView } from './licenses.view';
import { licensesDatatable } from './licenses-datatable';
import { licenseInfo } from './license-info';

export default angular
  .module('portainer.app.license-management.licenses-view', [])
  .component('licensesDatatable', licensesDatatable)
  .component('licenseInfo', licenseInfo)
  .component('licenseInfoPanel', LicenseInfoPanelAngular)
  .component('licensesView', licensesView).name;
