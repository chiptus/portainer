import { Role, User } from './types';

export function filterNonAdministratorUsers(users: User[]) {
  return users.filter((user) => !isPureAdmin(user));
}

type UserLike = Pick<User, 'Role'>;

// To avoid creating divergence between CE and EE
// isAdmin checks if the user is portainer admin or edge admin
export function isAdmin(user?: UserLike): boolean {
  return !!user && (user.Role === Role.Admin || user.Role === Role.EdgeAdmin);
}

// To avoid creating divergence between CE and EE
// isPureAdmin checks only if the user is portainer admin
// See bouncer.IsAdmin and bouncer.PureAdminAccess
export function isPureAdmin(user?: UserLike): boolean {
  return !!user && user.Role === Role.Admin;
}
