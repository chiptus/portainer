import angular from 'angular';

import { Header, HeaderAngular } from './Header';
import { HeaderContent, HeaderContentAngular } from './HeaderContent';
import { HeaderTitle, HeaderTitleAngular } from './HeaderTitle';
import { LicenseExpirationPanelAngular } from './LicenseExpirationPanel';

export { Header, HeaderTitle, HeaderContent };

export default angular
  .module('portainer.app.components.header', [])

  .component('licenseExpirationPanel', LicenseExpirationPanelAngular)
  .component('rdHeader', HeaderAngular)
  .component('rdHeaderContent', HeaderContentAngular)
  .component('rdHeaderTitle', HeaderTitleAngular).name;
