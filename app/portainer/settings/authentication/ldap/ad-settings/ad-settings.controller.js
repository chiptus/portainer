import _ from 'lodash-es';
import { FeatureId } from '@/portainer/feature-flags/enums';
const DEFAULT_GROUP_FILTER = '(objectClass=group)';
import { isLimitedToBE } from '@/portainer/feature-flags/feature-flags.service';

export default class AdSettingsController {
  /* @ngInject */
  constructor(LDAPService) {
    this.LDAPService = LDAPService;

    this.domainSuffix = '';
    this.limitedFeatureId = FeatureId.HIDE_INTERNAL_AUTH;
    this.onTlscaCertChange = this.onTlscaCertChange.bind(this);
    this.searchUsers = this.searchUsers.bind(this);
    this.searchGroups = this.searchGroups.bind(this);
    this.searchAdminGroups = this.searchAdminGroups.bind(this);
    this.parseDomainName = this.parseDomainName.bind(this);
    this.onAccountChange = this.onAccountChange.bind(this);
  }

  parseDomainName(account) {
    this.domainName = '';

    if (!account || !account.includes('@')) {
      return;
    }

    const [, domainName] = account.split('@');
    if (!domainName) {
      return;
    }

    const parts = _.compact(domainName.split('.'));
    this.domainSuffix = parts.map((part) => `dc=${part}`).join(',');
  }

  onAccountChange(account) {
    this.parseDomainName(account);
  }

  searchUsers() {
    return this.LDAPService.users(this.settings);
  }

  searchGroups() {
    return this.LDAPService.groups(this.settings);
  }

  searchAdminGroups() {
    if (this.settings.AdminAutoPopulate) {
      this.settings.AdminGroups = this.selectedAdminGroups;
    }

    const settings = {
      ...this.settings,
      AdminGroupSearchSettings: this.settings.AdminGroupSearchSettings.map((search) => ({ ...search, GroupFilter: search.GroupFilter || DEFAULT_GROUP_FILTER })),
    };

    return this.LDAPService.adminGroups(settings);
  }

  onTlscaCertChange(file) {
    this.tlscaCert = file;
  }

  addLDAPUrl() {
    this.settings.URLs.push('');
  }

  removeLDAPUrl(index) {
    this.settings.URLs.splice(index, 1);
  }

  isSaveSettingButtonDisabled() {
    return isLimitedToBE(this.limitedFeatureId) || !this.isLdapFormValid();
  }

  $onInit() {
    this.tlscaCert = this.settings.TLSCACert;
    this.parseDomainName(this.settings.ReaderDN);
  }
}
