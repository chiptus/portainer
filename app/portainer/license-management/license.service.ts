import {
  getLicenses,
  attachLicense,
  removeLicense,
  getLicenseInfo,
  unsubscribe,
  resetState,
  subscribe,
} from '@/react/portainer/licenses/license.service';

/* @ngInject */
export function LicenseService() {
  return {
    licenses: getLicenses,
    attach: attachLicense,
    remove: removeLicense,
    info: getLicenseInfo,
    subscribe,
    unsubscribe,
    resetState,
  };
}
