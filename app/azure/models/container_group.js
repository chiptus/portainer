import { AccessControlFormData } from 'Portainer/components/accessControlForm/porAccessControlFormModel';
import { ResourceControlViewModel } from 'Portainer/models/resourceControl/resourceControl';

export function ContainerGroupDefaultModel() {
  this.Location = '';
  this.OSType = 'Linux';
  this.Name = '';
  this.Image = '';
  this.AllocatePublicIP = true;
  this.Ports = [
    {
      container: 80,
      host: 80,
      protocol: 'TCP',
    },
  ];
  this.CPU = 1;
  this.Memory = 1;
  this.AccessControlData = new AccessControlFormData();
}

export function ContainerGroupViewModel(data) {
  const addressPorts = data.properties.ipAddress ? data.properties.ipAddress.ports : [];
  const container = data.properties.containers.length ? data.properties.containers[0] : {};
  const containerPorts = container ? container.properties.ports : [];

  this.Id = data.id;
  this.Name = data.name;
  this.Location = data.location;
  this.IPAddress = data.properties.ipAddress ? data.properties.ipAddress.ip : '';
  this.Ports = addressPorts.length
    ? addressPorts.map((binding, index) => {
        const port = containerPorts[index] ? containerPorts[index].port : undefined;
        return {
          container: port,
          host: binding.port,
          protocol: binding.protocol,
        };
      })
    : [];
  this.Image = container.properties.image || '';
  this.OSType = data.properties.osType;
  this.AllocatePublicIP = data.properties.ipAddress && data.properties.ipAddress.type === 'Public';
  this.CPU = container.properties.resources.requests.cpu;
  this.Memory = container.properties.resources.requests.memoryInGB;

  if (data.Portainer && data.Portainer.ResourceControl) {
    this.ResourceControl = new ResourceControlViewModel(data.Portainer.ResourceControl);
  }
}

export function CreateContainerGroupRequest(model) {
  this.location = model.Location;

  var containerPorts = [];
  var addressPorts = [];
  for (var i = 0; i < model.Ports.length; i++) {
    var binding = model.Ports[i];
    if (!binding.container || !binding.host) {
      continue;
    }

    containerPorts.push({
      port: binding.container,
    });

    addressPorts.push({
      port: binding.host,
      protocol: binding.protocol,
    });
  }

  this.properties = {
    osType: model.OSType,
    containers: [
      {
        name: model.Name,
        properties: {
          image: model.Image,
          ports: containerPorts,
          resources: {
            requests: {
              cpu: model.CPU,
              memoryInGB: model.Memory,
            },
          },
        },
      },
    ],
    ipAddress: {
      type: model.AllocatePublicIP ? 'Public' : 'Private',
      ports: addressPorts,
    },
  };
}
