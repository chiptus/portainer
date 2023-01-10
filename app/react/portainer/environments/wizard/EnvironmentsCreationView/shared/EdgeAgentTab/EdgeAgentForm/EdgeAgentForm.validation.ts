import { number, object, SchemaOf } from 'yup';

import { edgeAsyncIntervalsValidation } from '@/react/edge/components/EdgeAsyncIntervalsForm';
import { gpusListValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/Hardware/GpusList';
import { validation as urlValidation } from '@/react/portainer/common/PortainerUrlField';
import { validation as addressValidation } from '@/react/portainer/common/PortainerTunnelAddrField';

import { metadataValidation } from '../../MetadataFieldset/validation';
import { useNameValidation } from '../../NameField';

import { FormValues } from './types';

export function useValidationSchema(): SchemaOf<FormValues> {
  const nameValidation = useNameValidation();

  return object().shape({
    name: nameValidation,
    portainerUrl: urlValidation(),
    tunnelServerAddr: addressValidation(),
    pollFrequency: number().required(),
    meta: metadataValidation(),
    gpus: gpusListValidation(),
    edge: edgeAsyncIntervalsValidation(),
  });
}
