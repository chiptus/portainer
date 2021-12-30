import angular from 'angular';

import { Breadcrumbs } from './Breadcrumbs';
import { PageHeader } from './PageHeader';
import { HeaderContainer, HeaderAngular } from './HeaderContainer';
import { HeaderContent, HeaderContentAngular } from './HeaderContent';
import { HeaderTitle, HeaderTitleAngular } from './HeaderTitle';
import { LicenseExpirationPanelAngular } from './LicenseExpirationPanel';

export { PageHeader, Breadcrumbs, HeaderContainer, HeaderContent, HeaderTitle };

export default angular
  .module('portainer.app.components.header', [])

  .component('licenseExpirationPanel', LicenseExpirationPanelAngular)
  .component('rdHeader', HeaderAngular)
  .component('rdHeaderContent', HeaderContentAngular)
  .component('rdHeaderTitle', HeaderTitleAngular).name;
