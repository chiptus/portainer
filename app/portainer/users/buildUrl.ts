import { UserId } from './types';

export function buildUrl(id?: UserId, entity?: string) {
  let url = '/users';

  if (id) {
    url += `/${id}`;
  }

  if (entity) {
    url += `/${entity}`;
  }

  return url;
}
