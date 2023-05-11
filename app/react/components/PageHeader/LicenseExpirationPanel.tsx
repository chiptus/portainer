import moment from 'moment';
import clsx from 'clsx';
import { AlertTriangle } from 'lucide-react';

import { useLicenseInfo } from '@/react/portainer/licenses/use-license.service';
import { pluralize } from '@/portainer/helpers/strings';

import styles from './LicenseExpirationPanel.module.css';

export function LicenseExpirationPanelContainer() {
  const licenceInfoQuery = useLicenseInfo();

  if (licenceInfoQuery.isLoading || !licenceInfoQuery.info) {
    return null;
  }

  const nextLicenseExpiryUnix = moment.unix(
    licenceInfoQuery.info?.expiresAt || 0
  );
  const remainingDays = nextLicenseExpiryUnix.diff(
    moment().startOf('day'),
    'days'
  );
  const noValidLicense = !licenceInfoQuery.info?.valid;

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
