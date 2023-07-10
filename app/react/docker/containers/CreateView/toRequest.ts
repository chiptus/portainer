import { Registry } from '@/react/portainer/registries/types/registry';
import { buildImageFullURI } from '@/react/docker/images/utils';

import { convertToArrayOfStrings } from '@@/form-components/EnvironmentVariablesFieldset/utils';

import { baseFormUtils } from './BaseForm';
import { capabilitiesTabUtils } from './CapabilitiesTab';
import { commandsTabUtils } from './CommandsTab';
import { labelsTabUtils } from './LabelsTab';
import { networkTabUtils } from './NetworkTab';
import { resourcesTabUtils } from './ResourcesTab';
import { volumesTabUtils } from './VolumesTab';
import { CreateContainerRequest } from './types';
import { restartPolicyTabUtils } from './RestartPolicyTab';
import { Values } from './useInitialValues';

export function toRequest(values: Values, registry?: Registry) {
  let config: CreateContainerRequest = {
    Image: '',
    Env: [],
    MacAddress: '',
    ExposedPorts: {},
    WorkingDir: '',
    User: '',
    HostConfig: {
      RestartPolicy: {
        Name: 'no',
      },
      PortBindings: {},
      PublishAllPorts: false,
      Binds: [],
      AutoRemove: false,
      NetworkMode: 'bridge',
      Privileged: false,
      Init: false,
      ExtraHosts: [],
      Devices: [],
      DeviceRequests: [],
      CapAdd: [],
      CapDrop: [],
      Sysctls: {},
      LogConfig: {
        Type: 'none',
        Config: {},
      },
    },
    NetworkingConfig: {
      EndpointsConfig: {},
    },
    Labels: {},
  };

  config = commandsTabUtils.toRequest(config, values.commands);
  config = volumesTabUtils.toRequest(config, values.volumes);
  config = networkTabUtils.toRequest(config, values.network, '');
  config = labelsTabUtils.toRequest(config, values.labels);
  config = restartPolicyTabUtils.toRequest(config, values.restartPolicy);
  config = resourcesTabUtils.toRequest(config, values.resources);
  config = capabilitiesTabUtils.toRequest(config, values.capabilities);
  config = baseFormUtils.toRequest(config, values);
  config.Env = convertToArrayOfStrings(values.env);
  config.Image = buildImageFullURI(values.image.image, registry);

  return config;
}
