import { useEffect, useState } from 'react';
import moment from 'moment';
import clsx from 'clsx';
import { AlertTriangle } from 'lucide-react';

import { useLicenseInfo } from '@/react/portainer/licenses/use-license.service';
import { LicenseInfo } from '@/react/portainer/licenses/types';

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
  if (isLoading || !remainingDays || remainingDays >= 30) {
    return null;
  }

  const expirationMessage = buildMessage(remainingDays);

  return (
    <div className={clsx(styles.container)}>
      <div className={clsx(styles.item, 'vertical-center')}>
        <AlertTriangle className="icon icon-sm icon-warning" />
        <span className="text-muted">
          {expirationMessage} Please contact Portainer to renew your license.
        </span>
      </div>
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
  const { info, isLoading } = useLicenseInfo();

  const [remainingDays, setRemainingDays] = useState(0);

  useEffect(() => {
    if (info) {
      parseInfo(info);
    }
  }, [info]);

  return { remainingDays, isLoading };

  function parseInfo(info: LicenseInfo) {
    const expiresAt = moment.unix(info.expiresAt);
    setRemainingDays(expiresAt.diff(moment().startOf('day'), 'days'));
  }
}
