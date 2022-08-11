import moment from 'moment';

import { LicenseInfo, LicenseType } from '@/portainer/license-management/types';

export function getLicenseUpgradeURL(
  licenseInfo: LicenseInfo,
  usedNodes: number
) {
  let licenseType = 'trial';
  switch (licenseInfo.type) {
    case LicenseType.Trial:
      licenseType = 'trial';
      break;
    case LicenseType.Subscription:
      licenseType = 'subscription';
      break;
    case LicenseType.Essentials:
      licenseType = 'essentials';
      break;
    default:
      break;
  }

  const detail = JSON.stringify(licenseInfo);
  return `https://www.portainer.io/portainer-business-buy-more?used_nodes=${usedNodes}&type=${licenseType}&details=${detail}`;
}

export function calculateCountdownTime(enforcedAt: number) {
  const given = moment.unix(enforcedAt);
  const current = moment().startOf('day');

  const remaining = moment.duration(given.diff(current)).days();
  if (remaining < 0) {
    return '0';
  }

  return remaining;
}

export function getProductionEdition(edition: number) {
  let productionEdition = '';
  switch (edition) {
    case 1:
      productionEdition = 'Business Edition';
      break;
    case 2:
      productionEdition = 'Enterprise Edition';
      break;
    default:
      break;
  }
  return productionEdition;
}
