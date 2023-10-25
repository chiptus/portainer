import { boolean, object, SchemaOf, string } from 'yup';

import { validation as tunnelValidation } from '@/react/portainer/common/PortainerTunnelAddrField';
import { validation as urlValidation } from '@/react/portainer/common/PortainerUrlField';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';

import { file, withFileSize } from '@@/form-components/yup-file-validation';

import { FormValues } from './types';

const MAX_FILE_SIZE = 5_242_880; // 5MB

export function validationSchema(): SchemaOf<FormValues> {
  return object()
    .shape({
      EnableEdgeComputeFeatures: boolean().default(false),
      EnforceEdgeID: boolean().default(false),
    })
    .concat(
      isBE
        ? object({
            EdgePortainerUrl: urlValidation(),
            Edge: object({
              TunnelServerAddress: tunnelValidation(),
              MTLS: object({
                UseSeparateCert: boolean().default(false),
                CaCertFile: certValidation(),
                CertFile: certValidation(),
                KeyFile: certValidation(),
                CaCert: string().default(''),
                Cert: string().default(''),
                Key: string().default(''),
              }),
            }),
          })
        : object({
            EdgePortainerUrl: string().default(''),
            Edge: object({
              TunnelServerAddress: string().default(''),
              MTLS: object({
                UseSeparateCert: boolean().default(false),
                CaCertFile: file().notRequired(),
                CertFile: file().notRequired(),
                KeyFile: file().notRequired(),
                CaCert: string().notRequired(),
                Cert: string().notRequired(),
                Key: string().notRequired(),
              }),
            }),
          })
    );
}

function certValidation() {
  return withFileSize(file(), MAX_FILE_SIZE).when(['UseSeparateCert'], {
    is: (UseSeparateCert: boolean) => UseSeparateCert,
    then: (schema) => schema.optional(),
  });
}
