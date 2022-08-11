import { LicenseInfoPanel } from '@@/LicenseInfoPanel/LicenseInfoPanel';

import { useIntegratedLicenseInfo } from '../license-management/use-license.service';

export function LicenseNodePanel() {
  const integratedInfo = useIntegratedLicenseInfo();

  if (
    !integratedInfo ||
    integratedInfo.licenseInfo.enforcedAt === 0 ||
    integratedInfo.usedNodes <= integratedInfo.licenseInfo.nodes
  ) {
    return null;
  }

  const template = 'enforcement';

  return (
    <LicenseInfoPanel
      template={template}
      licenseInfo={integratedInfo.licenseInfo}
      usedNodes={integratedInfo.usedNodes}
    />
  );
}
