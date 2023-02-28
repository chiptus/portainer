import { LicenseInfoPanel } from '../licenses/components/LicenseInfoPanel';
import { useIntegratedLicenseInfo } from '../licenses/use-license.service';

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
