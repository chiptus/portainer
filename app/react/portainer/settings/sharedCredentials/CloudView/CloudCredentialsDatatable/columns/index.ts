import { buildNameColumn } from '@@/datatables/NameCell';

import { Credential } from '../../../types';

import { provider } from './provider';

export const columns = [
  buildNameColumn<Credential>(
    'name',
    'id',
    'portainer.settings.sharedcredentials.credential'
  ),
  provider,
];
