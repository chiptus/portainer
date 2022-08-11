import moment from 'moment';
import { Award, AlertCircle, ArrowUpRight } from 'react-feather';
import clsx from 'clsx';

import { LicenseInfo, LicenseType } from '@/portainer/license-management/types';
import { r2a } from '@/react-tools/react2angular';

import { ProgressBar } from '@@/ProgressBar';

import styles from './LicenseInfoPanel.module.css';
import {
  calculateCountdownTime,
  getLicenseUpgradeURL,
  getProductionEdition,
} from './utils';

interface Props {
  template: string;
  licenseInfo: LicenseInfo;
  usedNodes: number;
}

export function LicenseInfoPanel({ template, licenseInfo, usedNodes }: Props) {
  let widget = null;
  switch (template) {
    case 'info':
      widget = buildInfoWidget(licenseInfo, usedNodes);
      break;
    case 'alert':
      widget = buildAlertWidget(licenseInfo, usedNodes);
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

function buildInfoWidget(licenseInfo: LicenseInfo, usedNodes: number) {
  const contentSection = buildInfoContent(licenseInfo, usedNodes);
  const expiredAt = moment.unix(licenseInfo.expiresAt).format('YYYY-MM-DD');
  const extra = (
    <span className={styles.extra}>
      One or more your licenses will expire on <i>{expiredAt}</i>
    </span>
  );

  return (
    <div className={styles.licenseInfoPanel}>
      <div className={styles.licenseInfoContent}>
        {contentSection}
        <div>
          <b className="space-right">
            {usedNodes} / {licenseInfo.nodes} nodes used
          </b>

          <ProgressBar current={usedNodes} total={licenseInfo.nodes} />
        </div>
      </div>
      <hr className={styles.divider} />
      <div className={styles.extra}>
        <i
          className="fa fa-exclamation-triangle orange-icon"
          aria-hidden="true"
        />
        {extra}
      </div>
    </div>
  );
}

function buildAlertWidget(licenseInfo: LicenseInfo, usedNodes: number) {
  const contentSection = buildInfoContent(licenseInfo, usedNodes);
  const extra = (
    <span>
      You have exceeded the node allowance of your current license. Please
      contact your administrator.
    </span>
  );

  return (
    <div className={styles.licenseAlertPanel}>
      <div className={styles.licenseInfoContent}>
        {contentSection}
        <div>
          <b className="space-right">
            {usedNodes} / {licenseInfo.nodes} nodes used
          </b>

          <ProgressBar current={usedNodes} total={licenseInfo.nodes} />

          <div className={styles.alertExtra}>
            <i
              className="fa fa-exclamation-circle red-icon space-right"
              aria-hidden="true"
            />
            {extra}
          </div>
        </div>
      </div>
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
              {remainingTime} days remaining
            </span>
          </div>
          <div>
            <span>
              You have exceeded the node allowance of your license and will be
              unable to log into your account in {remainingTime} days. Please
              contact your administrator.
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
  const productionEdition = getProductionEdition(info.productEdition);
  const icon =
    usedNodes > info.nodes ? (
      <AlertCircle className="icon-danger icon-nested-red" aria-hidden="true" />
    ) : (
      <Award className="icon-primary icon-nested-blue" aria-hidden="true" />
    );

  const licenseUpgradeURL = getLicenseUpgradeURL(info, usedNodes);

  return (
    <div className={clsx(styles.licenseBlock)}>
      <div className={clsx(styles.icon)}>{icon}</div>

      <div className={clsx(styles.content)}>
        <div>
          <span className={clsx(styles.licenseTitle)}>License information</span>
        </div>
        <div className="control-label">
          Portainer {productionEdition} licensed to {info.company}
          {info.type !== LicenseType.Trial
            ? ` for up to ${info.nodes} ${info.nodes > 1 ? 'Nodes' : 'Node'}.`
            : null}
        </div>
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

export const LicenseInfoPanelAngular = r2a(LicenseInfoPanel, [
  'template',
  'licenseInfo',
  'usedNodes',
]);
