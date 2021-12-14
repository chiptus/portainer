import { useEffect, useState } from 'react';
import moment from 'moment';
import clsx from 'clsx';

import { r2a } from '@/react-tools/react2angular';
import { LicenseInfo } from '@/portainer/license-management/types';
import { useLicenseInfo } from '@/portainer/license-management/use-license.service';
import { error as notifyError } from '@/portainer/services/notifications';

import styles from './LicenseExpirationPanel.module.css';

export function LicenseExpirationPanelContainer() {
  const { remainingDays, isLoading } = useRemainingDays();

  return (
    <LicenseExpirationPanel
      remainingDays={remainingDays}
      isLoading={isLoading}
    />
  );
}

interface Props {
  remainingDays: number;
  isLoading?: boolean;
}

export function LicenseExpirationPanel({ remainingDays, isLoading }: Props) {
  if (isLoading || remainingDays >= 30) {
    return null;
  }

  const expirationMessage = buildMessage(remainingDays);

  return (
    <div className={clsx(styles.root, 'text-danger small')}>
      <i className="fa fa-exclamation-triangle space-right" />
      <strong className={styles.message}>
        <span className="space-right">{expirationMessage}</span>
        Please contact your administrator to renew your license.
      </strong>
    </div>
  );
}

function buildMessage(days: number) {
  return `One or more of your licenses ${expiringText(days)}.`;

  function expiringText(days: number) {
    if (days < 0) {
      return 'has expired';
    }

    if (days === 0) {
      return 'expires TODAY';
    }

    return `will expire in ${days === 1 ? '1 day' : `${days} days`}`;
  }
}

function useRemainingDays() {
  const { error, info, isLoading } = useLicenseInfo();

  useEffect(() => {
    if (error) {
      notifyError('Failure', error, 'Failed to get license info');
    }
  }, [error]);

  const [remainingDays, setRemainingDays] = useState(0);

  useEffect(() => {
    if (info) {
      parseInfo(info);
    }
  }, [info]);

  return { remainingDays, isLoading, error };

  function parseInfo(info: LicenseInfo) {
    const expiresAt = moment.unix(info.expiresAt);
    setRemainingDays(expiresAt.diff(moment().startOf('day'), 'days'));
  }
}

export const LicenseExpirationPanelAngular = r2a(
  LicenseExpirationPanelContainer,
  []
);
