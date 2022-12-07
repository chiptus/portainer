export default class LdapSettingsAdGroupSearchItemController {
  /* @ngInject */
  constructor($scope, Notifications) {
    Object.assign(this, { $scope, Notifications });

    this.groups = [];

    this.onChangeBaseDN = this.onChangeBaseDN.bind(this);

    this.onRemoveGroupBaseDN = this.onRemoveGroupBaseDN.bind(this);
    this.defaultGroupBaseDNRemoved = false;
    this.prevGroupBaseDN = '';
  }

  onChangeBaseDN(baseDN) {
    // Prevent deleted group base DN from being rewritten again
    if (this.defaultGroupBaseDNRemoved && this.prevGroupBaseDN === baseDN) {
      return;
    }
    this.config.GroupBaseDN = baseDN;
    this.prevGroupBaseDN = baseDN;
  }

  onRemoveGroupBaseDN() {
    return this.$scope.$evalAsync(() => {
      this.config.GroupBaseDN = '';
      this.defaultGroupBaseDNRemoved = true;
    });
  }

  addGroup() {
    this.groups.push({ type: 'ou', value: '' });
  }

  removeGroup($index) {
    this.groups.splice($index, 1);
    this.onGroupsChange();
  }

  onGroupsChange() {
    const groupsFilter = this.groups.map(({ type, value }) => `(${type}=${value})`).join('');
    this.onFilterChange(groupsFilter ? `(&${this.baseFilter}(|${groupsFilter}))` : `${this.baseFilter}`);
  }

  onFilterChange(filter) {
    this.config.GroupFilter = filter;
  }

  parseGroupFilter() {
    const match = this.config.GroupFilter.match(/^\(&\(objectClass=(\w+)\)\(\|((\(\w+=.+\))+)\)\)$/);
    if (!match) {
      return;
    }

    const [, , groupFilter = ''] = match;

    this.groups = groupFilter
      .slice(1, -1)
      .split(')(')
      .map((str) => str.split('='))
      .map(([type, value]) => ({ type, value }));
  }

  $onInit() {
    this.parseGroupFilter();
    this.defaultGroupBaseDNRemoved = false;
    this.prevGroupBaseDN = this.domainSuffix;
  }
}
