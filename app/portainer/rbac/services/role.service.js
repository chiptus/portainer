/* @ngInject */
export function RoleService($q, Roles) {
  return {
    roles,
  };

  function roles() {
    return Roles.query({}).$promise;
  }
}
