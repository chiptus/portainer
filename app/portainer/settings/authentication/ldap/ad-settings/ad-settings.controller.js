import _ from 'lodash-es';
import { FeatureId } from '@/react/portainer/feature-flags/enums';
const DEFAULT_GROUP_FILTER = '(objectClass=group)';
import { isLimitedToBE } from '@/react/portainer/feature-flags/feature-flags.service';

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
    this.onAccountChangeIfEmpty = this.onAccountChangeIfEmpty.bind(this);
  }

  parseDomainName(account) {
    this.domainName = '';

    if (!account || !account.includes('@')) {
      this.domainSuffix = '';
      return;
    }

    const [, domainName] = account.split('@');
    if (!domainName) {
      // When domainName does not exist anymore, we should clean up the
      // domainSuffix to overwrite the previous domainSuffix.
      // The issue happens when deleting the character with Backspace one by one
      this.domainSuffix = '';
      return;
    }

    const parts = _.compact(domainName.split('.'));
    this.domainSuffix = parts.map((part) => `dc=${part}`).join(',');
  }

  onAccountChange(account) {
    this.parseDomainName(account);
  }

  // ng-keyup event
  // This is helpful when user select all content and click Backspace to empty the field
  onAccountChangeIfEmpty(account) {
    if (account === '') {
      this.parseDomainName(account);
    }
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
