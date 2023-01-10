import { boolean, number, object, SchemaOf } from 'yup';

import { validation as tunnelValidation } from '@/react/portainer/common/PortainerTunnelAddrField';
import { validation as urlValidation } from '@/react/portainer/common/PortainerUrlField';

import { FormValues } from './types';

export function validationSchema(): SchemaOf<FormValues> {
  return object().shape({
    EnableEdgeComputeFeatures: boolean().required('This field is required.'),
    EnforceEdgeID: boolean().required('This field is required.'),
    EdgePortainerUrl: urlValidation(),
    Edge: object({
      TunnelServerAddress: tunnelValidation(),
    }),
    EdgeAgentCheckinInterval: number().default(0),
  });
}
