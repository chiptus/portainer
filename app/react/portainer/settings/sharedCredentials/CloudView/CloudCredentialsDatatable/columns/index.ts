import { buildNameColumn } from '@@/datatables/buildNameColumn';

import { Credential } from '../../../types';

import { provider } from './provider';

export const columns = [
  buildNameColumn<Credential>(
    'name',
    'portainer.settings.sharedcredentials.credential'
  ),
  provider,
];
