import { PortainerUrlField } from '@/react/portainer/common/PortainerUrlField';
import { PortainerTunnelAddrField } from '@/react/portainer/common/PortainerTunnelAddrField';

import { NameField } from '../../NameField';

interface EdgeAgentFormProps {
  readonly?: boolean;
}

export function EdgeAgentFieldset({ readonly }: EdgeAgentFormProps) {
  return (
    <>
      <NameField readonly={readonly} />
      <PortainerUrlField
        fieldName="portainerUrl"
        readonly={readonly}
        required
      />
      <PortainerTunnelAddrField
        fieldName="tunnelServerAddr"
        readonly={readonly}
        required
      />
    </>
  );
}
