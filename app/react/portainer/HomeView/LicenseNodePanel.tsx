import { useIntegratedLicenseInfo } from '@/portainer/license-management/use-license.service';

import { LicenseInfoPanel } from '@@/LicenseInfoPanel/LicenseInfoPanel';

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
