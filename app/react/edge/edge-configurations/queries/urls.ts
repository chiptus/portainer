import { EdgeConfiguration } from '../types';

export const BASE_URL = '/edge_configurations';

export function buildUrl({
  id,
  action,
}: {
  id?: EdgeConfiguration['id'];
  action?: string;
} = {}) {
  let url = BASE_URL;

  if (id) {
    url += `/${id}`;
  }

  if (action) {
    url += `/${action}`;
  }

  return url;
}
