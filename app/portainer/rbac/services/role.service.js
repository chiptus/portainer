const edgeAdminRole = {
  Id: 0,
  Name: 'Edge administrator',
  Description: 'Full control of all resources in all environments and access to the Edge Compute features',
  Priority: 0,
};

/* @ngInject */
export function RoleService($q, Roles) {
  return {
    roles,
  };

  async function roles() {
    const data = await Roles.query({}).$promise;
    // manually put the edge admin role as it is not an RBAC role in the backend
    // but is defined by user.Role = Role.EdgeAdministrator (3)
    data.push(edgeAdminRole);
    return data;
  }
}
