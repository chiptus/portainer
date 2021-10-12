/* @ngInject */
export function LDAPService(LDAP) {
  return { users, groups, adminGroups, check, testLogin };

  function users(ldapSettings) {
    return LDAP.users({ ldapSettings }).$promise;
  }

  async function groups(ldapSettings) {
    const userGroups = await LDAP.groups({ ldapSettings }).$promise;
    return userGroups.map(({ Name, Groups }) => {
      let name = Name;
      if (Name.includes(',') && Name.includes('=')) {
        const [cnName] = Name.split(',');
        const split = cnName.split('=');
        name = split[1];
      }
      return { Groups, Name: name };
    });
  }

  async function adminGroups(ldapSettings) {
    const userGroups = await LDAP.adminGroups({ ldapSettings }).$promise;
    return userGroups.sort((a, b) => (a.toLowerCase() > b.toLowerCase() ? 1 : -1)).map((name) => ({ name, selected: ldapSettings.AdminGroups.includes(name) }));
  }

  function check(ldapSettings) {
    return LDAP.check({ ldapSettings }).$promise;
  }

  function testLogin(ldapSettings, username, password) {
    return LDAP.testLogin({ ldapSettings, username, password }).$promise;
  }
}
