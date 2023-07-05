import moment from 'moment';
import { Award, AlertCircle, AlertTriangle, ArrowUpRight } from 'lucide-react';
import clsx from 'clsx';

import { Icon } from '@@/Icon';
import { ProgressBar } from '@@/ProgressBar';
import { Badge } from '@@/Badge';

import { LicenseInfo, LicenseType } from '../../types';

import { calculateCountdownTime, getLicenseUpgradeURL } from './utils';
import styles from './LicenseInfoPanel.module.css';

interface Props {
  template: 'info' | 'alert' | 'enforcement';
  licenseInfo: LicenseInfo;
  usedNodes: number;
  untrustedDevices?: number;
}

export function LicenseInfoPanel({
  template,
  licenseInfo,
  usedNodes,
  untrustedDevices = 0,
}: Props) {
  let widget = null;
  switch (template) {
    case 'info':
      widget = buildInfoWidget(licenseInfo, usedNodes, untrustedDevices);
      break;
    case 'alert':
      widget = buildAlertWidget(licenseInfo, usedNodes, untrustedDevices);
      break;
    case 'enforcement':
      widget = buildCountdownWidget(licenseInfo, usedNodes);
      break;
    default:
      break;
  }

  return (
    <div className="row">
      <div className="col-sm-12">{widget}</div>
    </div>
  );
}

function buildInfoWidget(
  licenseInfo: LicenseInfo,
  usedNodes: number,
  untrustedDevices: number
) {
  const contentSection = buildInfoContent(licenseInfo, usedNodes);
  const expiredAt = moment.unix(licenseInfo.expiresAt);
  const expiredAtStr = expiredAt.format('YYYY-MM-DD');
  const remainingDays = expiredAt.diff(moment().startOf('day'), 'days');

  let licenseExpiredInfo = (
    <div className={styles.extra}>
      <AlertTriangle className="icon icon-sm icon-warning" />
      <span className={styles.extraLessTwoMonths}>
        One or more of your licenses will expire on <i>{expiredAtStr}</i>
      </span>
    </div>
  );

  if (remainingDays >= 60) {
    licenseExpiredInfo = (
      <div className={styles.extra}>
        <span className="text-muted">
          One or more of your licenses will expire on <i>{expiredAtStr}</i>
        </span>
      </div>
    );
  }

  return (
    <div className={styles.licenseInfoPanel}>
      <div className={styles.licenseInfoContent}>
        {contentSection}
        <Details
          used={usedNodes}
          total={licenseInfo.nodes}
          valid={licenseInfo.valid}
          trial={licenseInfo.type === 1}
          untrustedDevices={untrustedDevices}
        />
      </div>
      <hr className={styles.divider} />
      {licenseExpiredInfo}
    </div>
  );
}

function buildAlertWidget(
  licenseInfo: LicenseInfo,
  usedNodes: number,
  untrustedDevices: number
) {
  const contentSection = buildInfoContent(licenseInfo, usedNodes);

  let exceededMsg =
    'You have exceeded the node allowance of your current license.';
  if (licenseInfo.nodes === 0) {
    exceededMsg = 'You have no current license or node allowance.';
  }

  return (
    <div className={styles.licenseAlertPanel}>
      <div>{contentSection}</div>
      <Details
        used={usedNodes}
        total={licenseInfo.nodes}
        valid={licenseInfo.valid}
        trial={licenseInfo.type === 1}
        untrustedDevices={untrustedDevices}
      />

      {licenseInfo.type !== LicenseType.Trial && (
        <div className={styles.alertExtra}>
          <Icon
            icon={AlertCircle}
            className={clsx('icon-danger', 'space-right')}
          />
          <span className={styles.alertExtraText}>
            {exceededMsg} Please contact Portainer to upgrade your license.
          </span>
        </div>
      )}
    </div>
  );
}

function buildCountdownWidget(licenseInfo: LicenseInfo, usedNodes: number) {
  const licenseUpgradeURL = getLicenseUpgradeURL(licenseInfo, usedNodes);
  const remainingTime = calculateCountdownTime(licenseInfo.enforcedAt);

  return (
    <div className={styles.licenseHomeInfo}>
      <div className={styles.licenseBlock}>
        <div className={clsx(styles.icon)}>
          <AlertCircle
            className="icon-danger icon-nested-pink"
            aria-hidden="true"
          />
        </div>

        <div className={clsx(styles.content)}>
          <div>
            <span className={styles.licenseTitle}>
              {remainingTime} remaining
            </span>
          </div>
          <div>
            <span>
              You have exceeded the node allowance of your license and your
              users will be unable to log into their accounts in {remainingTime}
              . Please contact Portainer to upgrade your license.
            </span>
          </div>
        </div>

        <div className={clsx(styles.button)}>
          <a
            className="btn btn-primary btn-sm"
            href={licenseUpgradeURL}
            target="_blank"
            rel="noreferrer"
          >
            Buy or renew licenses
          </a>
        </div>
      </div>
    </div>
  );
}

function buildInfoContent(info: LicenseInfo, usedNodes: number) {
  const icon =
    usedNodes > info.nodes && info.type !== LicenseType.Trial ? (
      <AlertCircle className="icon-danger icon-nested-red" aria-hidden="true" />
    ) : (
      <Award className="icon-primary icon-nested-blue" aria-hidden="true" />
    );

  const licenseUpgradeURL = getLicenseUpgradeURL(info, usedNodes);

  let subtitle = (
    <div className="control-label">
      Portainer licensed to {info.company}
      {info.type !== LicenseType.Trial
        ? ` for up to ${info.nodes} ${info.nodes !== 1 ? 'nodes' : 'node'}.`
        : null}
    </div>
  );
  if (!info.valid) {
    subtitle = (
      <div className="control-label">Portainer has no current license.</div>
    );
  }

  return (
    <div className={clsx(styles.licenseBlock)}>
      <div className={clsx(styles.icon)}>{icon}</div>

      <div className={clsx(styles.content)}>
        <div>
          <span className={clsx(styles.licenseTitle)}>License information</span>
        </div>
        {subtitle}
      </div>

      <div className={clsx(styles.button)}>
        <a
          className="btn btn-primary btn-sm"
          href={licenseUpgradeURL}
          target="_blank"
          rel="noreferrer"
        >
          <ArrowUpRight aria-hidden="true" size={12} />
          Upgrade licenses
        </a>
      </div>
    </div>
  );
}

function Details({
  used,
  total,
  valid,
  trial,
  untrustedDevices,
}: {
  used: number;
  total: number;
  valid: boolean;
  trial: boolean;
  untrustedDevices: number;
}) {
  let nodesUsedMsg = `${used.toString()} / ${total.toString()} nodes used`;
  if (total === 0 && valid) {
    nodesUsedMsg = `${used.toString()} out of unlimited nodes used`;
  }
  let color = used > total ? '#f04438' : '#0086c9';
  if (trial) {
    color = '#0086c9';
  }

  return (
    <div>
      <div className="flex">
        <b className="space-right">{nodesUsedMsg}</b>
        <Badge type={used + untrustedDevices > total ? 'warn' : 'info'}>
          +{untrustedDevices} in waiting room
        </Badge>
      </div>

      <ProgressBar
        steps={[
          {
            value: used,
            color,
          },
          {
            value: untrustedDevices,
            color:
              used + untrustedDevices > total
                ? 'var(--ui-warning-3)'
                : 'var(--ui-blue-3)',
          },
        ]}
        total={total}
      />
    </div>
  );
}
