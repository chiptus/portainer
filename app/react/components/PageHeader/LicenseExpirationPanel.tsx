import moment from 'moment';
import clsx from 'clsx';
import { AlertTriangle } from 'lucide-react';

import { useLicensesQuery } from '@/react/portainer/licenses/use-license.service';
import { License } from '@/react/portainer/licenses/types';
import { pluralize } from '@/portainer/helpers/strings';

import styles from './LicenseExpirationPanel.module.css';

export function LicenseExpirationPanelContainer() {
  const licensesQuery = useLicensesQuery();
  if (licensesQuery.isLoading || !licensesQuery.data) {
    return null;
  }

  const { remainingDays, noValidLicense } = getDaysUntilNextLicenseExpiry(
    licensesQuery.data
  );

  return (
    <LicenseExpirationPanel
      remainingDays={remainingDays}
      noValidLicense={noValidLicense}
    />
  );
}

interface Props {
  remainingDays: number;
  noValidLicense?: boolean;
}

export function LicenseExpirationPanel({
  remainingDays,
  noValidLicense,
}: Props) {
  if (remainingDays > 30) {
    return null;
  }

  const expirationMessage = buildMessage(remainingDays, noValidLicense);

  return (
    <div className={clsx(styles.container)}>
      <div className={clsx(styles.item, 'vertical-center')}>
        <AlertTriangle className="icon icon-sm icon-warning shrink-0" />
        <span className="text-muted">{expirationMessage}</span>
      </div>
    </div>
  );
}

function buildMessage(days: number, noValidLicense?: boolean) {
  if (noValidLicense) {
    return 'You have no valid licenses and will need to supply one on next login. Please contact Portainer to purchase a license.';
  }

  return `One or more of your licenses ${expiringText(
    days
  )}. Please contact Portainer to renew your license.`;

  function expiringText(days: number) {
    if (days < 0) {
      return 'has expired';
    }

    if (days === 0) {
      return 'expires TODAY';
    }

    return `will expire in ${days} ${pluralize(days, 'day')}`;
  }
}

// getDaysUntilNextLicenseExpiry gets the remaining days for the license that's the closest to expire.
function getDaysUntilNextLicenseExpiry(licenses: License[]) {
  const licensesExpiries = licenses.map((license) => license.expiresAt);
  // filter out expired licenses
  const filteredLicensesExpiries = licensesExpiries.filter(
    (expiresAt) => expiresAt > moment().unix()
  );
  // if there are no valid licenses, return noValidLicense: true, remainingDays: 0
  if (filteredLicensesExpiries.length === 0) {
    return {
      noValidLicense: true,
      remainingDays: 0,
    };
  }
  const nextLicenseExpiryUnix = Math.min(...filteredLicensesExpiries);
  const nextLicenseExpiryTime = moment.unix(nextLicenseExpiryUnix);
  const daysUntilNextLicenseExpiry = nextLicenseExpiryTime.diff(
    moment().startOf('day'),
    'days'
  );
  return {
    noValidLicense: false,
    remainingDays: daysUntilNextLicenseExpiry,
  };
}
