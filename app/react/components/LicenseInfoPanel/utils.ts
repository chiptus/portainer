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
  const currentHour = moment().startOf('hour');
  const totalHours = Math.trunc(
    moment.duration(given.diff(currentHour)).asHours()
  );

  const remainingDay = Math.trunc(totalHours / 24);
  const remainingHour = totalHours % 24;
  if (remainingDay === 0 && remainingHour > 0) {
    return pluralizeTimeUnit(remainingHour, 'hour');
  }
  if (remainingHour === 0 && remainingDay > 0) {
    return pluralizeTimeUnit(remainingDay, 'day');
  }
  if (remainingHour > 0 && remainingDay > 0) {
    return `${pluralizeTimeUnit(remainingDay, 'day')} and ${pluralizeTimeUnit(
      remainingHour,
      'hour'
    )}`;
  }
  return '0 days';
}

function pluralizeTimeUnit(value: number, unit: string) {
  if (value > 1) {
    return `${value} ${unit}s`;
  }
  if (value > 0 && value <= 1) {
    return `${value} ${unit}`;
  }
  return `${0} ${unit}`;
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
