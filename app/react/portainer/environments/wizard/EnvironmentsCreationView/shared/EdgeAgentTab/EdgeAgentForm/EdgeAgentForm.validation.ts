import { number, object, SchemaOf, string } from 'yup';

import { edgeAsyncIntervalsValidation } from '@/react/edge/components/EdgeAsyncIntervalsForm';
import { validation as urlValidation } from '@/react/portainer/common/PortainerUrlField';
import { validation as addressValidation } from '@/react/portainer/common/PortainerTunnelAddrField';

import { metadataValidation } from '../../MetadataFieldset/validation';
import { useNameValidation } from '../../NameField';

import { FormValues } from './types';

export function useValidationSchema(asyncMode: boolean): SchemaOf<FormValues> {
  const nameValidation = useNameValidation();

  return object().shape({
    name: nameValidation,
    portainerUrl: urlValidation(),
    tunnelServerAddr: asyncMode ? string() : addressValidation(),
    pollFrequency: number().required(),
    meta: metadataValidation(),
    edge: edgeAsyncIntervalsValidation(),
  });
}
