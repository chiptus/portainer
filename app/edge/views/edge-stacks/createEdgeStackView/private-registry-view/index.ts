import { react2angular } from '@/react-tools/react2angular';

import { PrivateRegistryView } from './PrivateRegistryView';

export const PrivateRegistryViewAngular = react2angular(PrivateRegistryView, [
  'value',
  'registries',
  'onChange',
  'forminvalid',
  'errorMessage',
  'onSelect',
  'isActive',
  'clearRegistries',
]);
